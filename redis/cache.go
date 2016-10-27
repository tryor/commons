//cache
package redis

import (
	"encoding/json"
	//	"fmt"
	//	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

var pool *redis.Pool

func CacheInit(server, password string) {
	pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func GetRedis() redis.Conn {
	return pool.Get()
}

func Send(cmd string, args ...interface{}) error {
	red := GetRedis()
	defer red.Close()
	err := red.Send(cmd, args...)
	//log.Debugf("Send %v %v %v", cmd, args, err)
	if err != nil {
		return err
	}
	return red.Flush()
}

func Do(cmd string, args ...interface{}) (interface{}, error) {
	red := GetRedis()
	defer red.Close()
	//log.Debugf("Do %v %v", cmd, args)
	return red.Do(cmd, args...)
}

func SetString(k string, v string, expire ...int) error {
	//SET key value [EX seconds]
	if len(expire) > 0 {
		return Send("SET", k, v, "EX", expire[0]) //return Send("SET", k, v, "EX "+strconv.Itoa(expire[0]))
	} else {
		return Send("SET", k, v)
	}
}

func GetString(k string) (string, error) {
	str, err := redis.String(Do("GET", k))
	if err == nil {
		str = strings.Trim(str, "\"")
	}
	return str, err
}

func SetObject(k string, v interface{}, expire ...int) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if len(expire) > 0 {
		return Send("SET", k, b, "EX", expire[0])
	} else {
		return Send("SET", k, b)
	}
}

func GetObject(k string, clazz interface{}) error {
	b, err := redis.Bytes(Do("GET", k))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, clazz)
}

func Del(k string) error {
	return Send("DEL", k)
}

func titleCasedName(name string) string {
	newstr := make([]rune, 0)
	upNextChar := true

	for _, chr := range name {
		switch {
		case upNextChar:
			upNextChar = false
			chr -= ('a' - 'A')
		case chr == '_':
			upNextChar = true
			continue
		}

		newstr = append(newstr, chr)
	}

	return string(newstr)
}

type HashMap struct {
	Name string
}

func NewHashMap(name string, expire ...int) *HashMap {
	Do("PING")
	hmap := &HashMap{name}
	if len(expire) > 0 {
		hmap.SetExpire(expire[0])
	}
	return hmap
}

func (this *HashMap) SetExpire(second int) error {
	return Send("EXPIRE", this.Name, second)
}

func (this *HashMap) SetObject(k string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return Send("HSET", this.Name, k, b)
}

func (this *HashMap) GetObject(k string, clazz interface{}) error {
	b, err := redis.Bytes(Do("HGET", this.Name, k))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, clazz)
}

//func (orm *HashMap) ScanPK(output interface{}) *Model {
//	if reflect.TypeOf(reflect.Indirect(reflect.ValueOf(output)).Interface()).Kind() == reflect.Slice {
//		sliceValue := reflect.Indirect(reflect.ValueOf(output))
//		sliceElementType := sliceValue.Type().Elem()
//		for i := 0; i < sliceElementType.NumField(); i++ {
//			bb := reflect.ValueOf(sliceElementType.Field(i).Tag)
//			if bb.String() == "PK" {
//				orm.PrimaryKey = sliceElementType.Field(i).Name
//			}
//		}
//	} else {
//		tt := reflect.TypeOf(reflect.Indirect(reflect.ValueOf(output)).Interface())
//		for i := 0; i < tt.NumField(); i++ {
//			bb := reflect.ValueOf(tt.Field(i).Tag)
//			if bb.String() == "PK" {
//				orm.PrimaryKey = tt.Field(i).Name
//			}
//		}
//	}
//	return orm

//}

//func (this *HashMap) GetObjectList(k []string, objs []interface{}) error {
//	args := []interface{}{}
//	args = append(args, this.Name)
//	for _, v := range k {
//		args = append(args, v)
//	}
//	b, err := redis.MultiBulk(Do("HMGET", args...))
//	if err != nil {
//		return err
//	}
//	for i, v := range b {
//		bb, err := redis.Bytes(v, nil)
//		if err != nil {
//			break
//		}
//		err = json.Unmarshal(bb, objs[i])
//		if err != nil {
//			break
//		}
//	}
//	return err
//}

func (this *HashMap) SetString(k string, v string) error {
	return Send("HSET", this.Name, k, v)
}

func (this *HashMap) GetString(k string) (string, error) {
	str, err := redis.String(Do("HGET", this.Name, k))
	if err == nil {
		str = strings.Trim(str, "\"")
	}
	return str, err
}

func (this *HashMap) GetStringList(k []string) ([]string, error) {
	args := []interface{}{}
	args = append(args, this.Name)
	for _, v := range k {
		args = append(args, v)
	}
	reply, err := redis.MultiBulk(Do("HMGET", args...))
	if err != nil {
		return nil, err
	}
	var list = make([]string, 0)
	for _, v := range reply {
		s, err := redis.String(v, nil)
		if err != nil {
			break
		}
		s = strings.Trim(s, "\"")
		list = append(list, s)
	}
	return list, err
}

func (this *HashMap) Size() (int, error) {
	return redis.Int(Do("HLEN", this.Name))
}

func (this *HashMap) Del(k string) error {
	return Send("HDEL", this.Name, k)
}

func (this *HashMap) Exists(k string) bool {
	v, err := redis.Bool(Do("HEXISTS", this.Name, k))
	if err != nil {
		return false
	}
	return v
}

func (this *HashMap) Clear() error {
	return Send("DEL", this.Name)
}

type SortedSet struct {
	Name string
}

func NewSortedSet(name string) *SortedSet {
	Do("PING")
	return &SortedSet{name}
}

func (this *SortedSet) SetExpire(second int) error {
	return Send("EXPIRE", this.Name, second)
}

func (this *SortedSet) AddObject(score float64, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return Send("ZADD", this.Name, score, b)
}

func (this *SortedSet) AddString(score float64, v string) error {
	return Send("ZADD", this.Name, score, v)
}

func (this *SortedSet) Size() int {
	b, err := redis.Int(Do("ZCARD", this.Name))
	if err != nil {
		return -1
	}
	return b
}

func (this *SortedSet) SizeByScore(min, max float64) int {
	b, err := redis.Int(Do("ZCOUNT", this.Name, min, max))
	if err != nil {
		return -1
	}
	return b
}

func (this *SortedSet) GetObject(index int, clazz interface{}) error {
	b, err := redis.Bytes(Do("ZRANGE", this.Name, index, index+1))
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, clazz)
	return err
}

//func (this *SortedSet) GetObjects(clazz []interface{}, start, limit int) error {
//	b, err := redis.MultiBulk(Do("ZRANGE", this.Name, start, start+limit))
//	if err != nil {
//		return err
//	}
//	for i, v := range b {
//		bb, err := redis.Bytes(v, nil)
//		if err != nil {
//			break
//		}
//		err = json.Unmarshal(bb, &clazz[i])
//		if err != nil {
//			break
//		}
//	}
//	return err
//}

func (this *SortedSet) RemoveObject(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return Send("ZREM", this.Name, b)
}

func (this *SortedSet) GetString(index int) (string, error) {
	str, err := redis.String(Do("ZRANGE", this.Name, index, index+1))
	if err == nil {
		str = strings.Trim(str, "\"")
	}
	return str, err
}

func (this *SortedSet) GetStrings(start, limit int) ([]string, error) {
	a, err := Do("ZRANGE", this.Name, start, start+limit)
	if err != nil {
		return nil, err
	}
	b, err := redis.MultiBulk(a, err)
	if err != nil {
		return nil, err
	}

	var list = make([]string, 0)
	for _, v := range b {
		s, err := redis.String(v, nil)
		if err != nil {
			break
		}
		s = strings.Trim(s, "\"")
		list = append(list, s)
	}
	return list, err
}

func (this *SortedSet) GetStringsRev(start, limit int) ([]string, error) {
	a, err := Do("ZREVRANGE", this.Name, start, start+limit)
	if err != nil {
		return nil, err
	}
	b, err := redis.MultiBulk(a, err)
	if err != nil {
		return nil, err
	}

	var list = make([]string, 0)
	for _, v := range b {
		s, err := redis.String(v, nil)
		if err != nil {
			break
		}
		s = strings.Trim(s, "\"")
		list = append(list, s)
	}
	return list, err
}

func (this *SortedSet) RemoveString(v string) error {
	return Send("ZREM", this.Name, v)
}

func (this *SortedSet) Remove(start, limit int) error {
	return Send("ZREMRANGEBYRANK", this.Name, start, start+limit-1)
}

func (this *SortedSet) RemoveByIndex(index int) error {
	return Send("ZREMRANGEBYRANK", this.Name, index, index)
}

func (this *SortedSet) ObjectScore(v interface{}) int {
	b, err := json.Marshal(v)
	if err != nil {
		return -1
	}
	r, err := redis.Int(Do("ZINCRBY", this.Name, b))
	if err != nil {
		return -1
	}
	return r
}

func (this *SortedSet) StringScore(v string) int {
	r, err := redis.Int(Do("ZINCRBY", this.Name, v))
	if err != nil {
		return -1
	}
	return r
}

func (this *SortedSet) Clear() error {
	return Send("DEL", this.Name)
}
