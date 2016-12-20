package cache

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

type redisCache struct {
	p               *redis.Pool
	addr            string
	dbNum           int
	password        string
	maxIdleConns    int
	connIdleTimeout int //秒, 连接空闲超时时间
	//获取连接时，免ping时间，即此连接connIdleTimeout时间如果小于noTesttime，将不ping
	noTesttime    time.Duration
	defaultExpire int64 //秒，数据默认过期时间
}

func NewRedisCache() Cache {
	return &redisCache{}
}

func (rc *redisCache) NewMap(name string, expire ...time.Duration) (Map, error) {
	var timeout int64
	if len(expire) > 0 {
		timeout = int64(expire[0] / time.Second)
	} else if rc.defaultExpire > 0 {
		timeout = rc.defaultExpire
	}

	return &redisMap{rc: rc, name: name, expire: timeout}, nil
}

func (rc *redisCache) SetExpire(key string, expire ...time.Duration) error {
	if len(expire) > 0 {
		return rc.send("EXPIRE", key, int64(expire[0]/time.Second))
	}
	return nil
}

func (rc *redisCache) Put(key string, val string, expire ...time.Duration) error {
	if len(expire) > 0 {
		return rc.send("SETEX", key, int64(expire[0]/time.Second), val)
	} else if rc.defaultExpire > 0 {
		return rc.send("SETEX", key, rc.defaultExpire, val)
	} else {
		return rc.send("SET", key, val)
	}
}

func redisString(data interface{}, err error) (ret string, rerr error) {
	ret, rerr = redis.String(data, err)
	if rerr == redis.ErrNil {
		rerr = ErrNil
	}
	return
}

func (rc *redisCache) Get(key string) (str string, err error) {
	str, err = redisString(rc.do("GET", key))
	return
}
func (rc *redisCache) GetMulti(keys []string) ([]string, error) {
	args := []interface{}{}
	for _, v := range keys {
		args = append(args, v)
	}
	reply, err := redis.MultiBulk(rc.do("MGET", args...))
	if err != nil {
		return nil, err
	}
	var list = make([]string, 0)
	for _, v := range reply {
		s, err1 := redisString(v, nil)
		if err1 != nil && err1 != ErrNil {
			err = err1
		}
		list = append(list, s)
	}
	return list, err
}

func (rc *redisCache) PutObject(key string, val interface{}, expire ...time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	if len(expire) > 0 {
		return rc.send("SETEX", key, int64(expire[0]/time.Second), b)
	} else if rc.defaultExpire > 0 {
		return rc.send("SETEX", key, rc.defaultExpire, b)
	} else {
		return rc.send("SET", key, b)
	}
}

func redisBytes(data interface{}, err error) (ret []byte, rerr error) {
	ret, rerr = redis.Bytes(data, err)
	if rerr == redis.ErrNil {
		rerr = ErrNil
	}
	return
}

func (rc *redisCache) GetObject(key string, objptr interface{}) error {
	b, err := redisBytes(rc.do("GET", key))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, objptr)
}

func (rc *redisCache) Delete(key string) error {
	return rc.send("DEL", key)
}

func redisBool(data interface{}, err error) (ret bool, rerr error) {
	ret, rerr = redis.Bool(data, err)
	if rerr == redis.ErrNil {
		rerr = ErrNil
	}
	return
}

func (rc *redisCache) Incr(key string) error {
	_, err := redisBool(rc.do("INCRBY", key, 1))
	return err
}
func (rc *redisCache) Decr(key string) error {
	_, err := redisBool(rc.do("INCRBY", key, -1))
	return err
}
func (rc *redisCache) Exists(key string) bool {
	v, err := redis.Bool(rc.do("EXISTS", key))
	if err != nil {
		return false
	}
	return v
}

func (rc *redisCache) send(cmd string, args ...interface{}) error {
	red := rc.p.Get()
	defer red.Close()
	err := red.Send(cmd, args...)
	if err != nil {
		return err
	}
	return red.Flush()
}

func (rc *redisCache) do(cmd string, args ...interface{}) (interface{}, error) {
	red := rc.p.Get()
	defer red.Close()
	return red.Do(cmd, args...)
}

//config - {"addr":"", "password":"", "dbNum":"0", "maxIdleConns":"10", "connIdleTimeout":"300", "noTesttime":"60", "defaultExpire":""}
//defaultExpire - 默认过期时间，秒
func (rc *redisCache) Init(config string) error {
	var cf map[string]string
	json.Unmarshal([]byte(config), &cf)

	if _, ok := cf["addr"]; !ok {
		return errors.New("config has no addr key")
	}
	if _, ok := cf["dbNum"]; !ok {
		cf["dbNum"] = "0"
	}
	if _, ok := cf["password"]; !ok {
		cf["password"] = ""
	}

	if _, ok := cf["maxIdleConns"]; !ok {
		cf["maxIdleConns"] = "3"
	}

	if _, ok := cf["connIdleTimeout"]; !ok {
		cf["connIdleTimeout"] = "300"
	}

	if _, ok := cf["noTesttime"]; !ok {
		cf["noTesttime"] = "60"
	}

	if _, ok := cf["defaultExpire"]; !ok {
		cf["defaultExpire"] = "0"
	}

	rc.addr = cf["addr"]
	rc.dbNum, _ = strconv.Atoi(cf["dbNum"])
	rc.password = cf["password"]
	rc.maxIdleConns, _ = strconv.Atoi(cf["maxIdleConns"])
	rc.connIdleTimeout, _ = strconv.Atoi(cf["connIdleTimeout"])
	noTesttime, _ := strconv.Atoi(cf["noTesttime"])
	rc.noTesttime = time.Duration(noTesttime) * time.Second
	rc.defaultExpire, _ = strconv.ParseInt(cf["defaultExpire"], 10, 0)

	rc.connectInit()

	c := rc.p.Get()
	defer c.Close()

	return c.Err()
}

func (rc *redisCache) connectInit() {
	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial("tcp", rc.addr)
		if err != nil {
			return nil, err
		}

		if rc.password != "" {
			if _, err := c.Do("AUTH", rc.password); err != nil {
				c.Close()
				return nil, err
			}
		}

		_, selecterr := c.Do("SELECT", rc.dbNum)
		if selecterr != nil {
			c.Close()
			return nil, selecterr
		}
		return
	}

	testOnBorrow := func(c redis.Conn, t time.Time) error {
		if time.Now().Sub(t) < rc.noTesttime {
			return nil
		}
		_, err := c.Do("PING")
		if err != nil {
			return err
		}
		return nil
	}

	// initialize a new pool
	rc.p = &redis.Pool{
		MaxIdle:      rc.maxIdleConns,
		IdleTimeout:  time.Duration(rc.connIdleTimeout) * time.Second,
		Dial:         dialFunc,
		TestOnBorrow: testOnBorrow,
	}
}

type redisMap struct {
	rc     *redisCache
	name   string
	expire int64
}

func (m *redisMap) put(key string, val interface{}) error {
	var err error
	if m.expire == 0 {
		err = m.rc.send("HSET", m.name, key, val)
	} else {
		mapexist := m.rc.Exists(m.name)
		err = m.rc.send("HSET", m.name, key, val)
		if err == nil {
			if !mapexist {
				err = m.rc.send("EXPIRE", m.name, m.expire)
			}
		}
	}
	return err
}

func (m *redisMap) Put(key string, val string) error {
	return m.put(key, val)
}

func (m *redisMap) Get(key string) (string, error) {
	return redisString(m.rc.do("HGET", m.name, key))
}
func (m *redisMap) GetMulti(keys []string) ([]string, error) {
	args := []interface{}{}
	args = append(args, m.name)
	for _, v := range keys {
		args = append(args, v)
	}
	reply, err := redis.MultiBulk(m.rc.do("HMGET", args...))
	if err != nil {
		return nil, err
	}
	var list = make([]string, 0)
	for _, v := range reply {
		s, err1 := redisString(v, nil)
		if err1 != nil && err1 != ErrNil {
			err = err1
		}
		//s = strings.Trim(s, "\"")
		list = append(list, s)
	}
	return list, err
}

func (m *redisMap) PutObject(key string, val interface{}) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return m.put(key, b)
}
func (m *redisMap) GetObject(key string, valptr interface{}) error {
	b, err := redisBytes(m.rc.do("HGET", m.name, key))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, valptr)

}

func (m *redisMap) Delete(key string) error {
	return m.rc.send("HDEL", m.name, key)
}

func (m *redisMap) Incr(key string) error {
	var err error
	if m.expire == 0 {
		_, err = redisBool(m.rc.do("HINCRBY", m.name, key, 1))
	} else {
		mapexist := m.rc.Exists(m.name)
		_, err = redisBool(m.rc.do("HINCRBY", m.name, key, 1))
		if err == nil {
			if !mapexist {
				err = m.rc.send("EXPIRE", m.name, m.expire)
			}
		}
	}
	return err
}

func (m *redisMap) Decr(key string) error {
	var err error
	if m.expire == 0 {
		_, err = redisBool(m.rc.do("HINCRBY", m.name, key, -1))
	} else {
		mapexist := m.rc.Exists(m.name)
		_, err = redisBool(m.rc.do("HINCRBY", m.name, key, -1))
		if err == nil {
			if !mapexist {
				err = m.rc.send("EXPIRE", m.name, m.expire)
			}
		}
	}
	return err
}

func (m *redisMap) Exists(key string) bool {
	v, err := redis.Bool(m.rc.do("HEXISTS", m.name, key))
	if err != nil {
		return false
	}
	return v
}

func (m *redisMap) Size() (int, error) {
	return redis.Int(m.rc.do("HLEN", m.name))
}

func (m *redisMap) Clear() error {
	return m.rc.send("DEL", m.name)
}

func init() {
	Register("redis", NewRedisCache)
}
