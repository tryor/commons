package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type memoryEntry struct {
	value       interface{}
	createdtime time.Time
	expire      time.Duration
}

func (e *memoryEntry) isExpire() bool {
	if e.expire == 0 {
		return false
	}
	return time.Now().Sub(e.createdtime) > e.expire
}

type memoryCache struct {
	lock          *sync.RWMutex
	gccyc         time.Duration
	items         map[string]*memoryEntry
	defaultExpire time.Duration //数据默认过期时间
}

func NewMemoryCache() Cache {
	return &memoryCache{items: make(map[string]*memoryEntry), lock: new(sync.RWMutex)}
}

//config - {"gccyc":60, "defaultExpire":10}, second
//gccyc - GC周期， 秒， 默认：60
//defaultExpire - 默认过期时间，秒， 默认：0, 数据将不会过期
func (c *memoryCache) Init(config string) error {
	cf := make(map[string]int)
	json.Unmarshal([]byte(config), &cf)
	if _, ok := cf["gccyc"]; !ok {
		cf["gccyc"] = 60
	}

	if _, ok := cf["defaultExpire"]; !ok {
		cf["defaultExpire"] = 0
	}

	c.gccyc = time.Duration(cf["gccyc"]) * time.Second
	c.defaultExpire = time.Duration(cf["defaultExpire"]) * time.Second

	go c.gc()

	return nil
}

func (c *memoryCache) gc() {
	if (c.gccyc / time.Second) < 1 {
		return
	}
	for {
		<-time.After(c.gccyc)
		if c.items == nil {
			return
		}

		c.itemExpireds()
		//for name := range c.items {
		//	c.itemExpired(name)
		//}
	}
}

func (c *memoryCache) itemExpireds() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for key, itm := range c.items {
		if itm.isExpire() {
			delete(c.items, key)
			return
		}
	}
}

func (c *memoryCache) itemExpired(name string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	itm, ok := c.items[name]
	if !ok {
		return true
	}
	if itm.isExpire() {
		delete(c.items, name)
		return true
	}
	return false
}

func (c *memoryCache) put(key string, val interface{}, expire ...time.Duration) *memoryEntry {

	var timeout time.Duration
	if len(expire) > 0 {
		timeout = expire[0]
	} else if c.defaultExpire > 0 {
		timeout = c.defaultExpire
	}

	itm := &memoryEntry{
		value:       val,
		createdtime: time.Now(),
		expire:      timeout,
	}
	c.items[key] = itm
	return itm
}

func (c *memoryCache) Put(key string, val string, expire ...time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.put(key, val, expire...)
	return nil
}

func (c *memoryCache) lget(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if itm, ok := c.items[key]; !ok || itm.isExpire() {
		return nil
	} else {
		return itm
	}
}

func (c *memoryCache) Get(key string) (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if itm, ok := c.items[key]; ok {
		if itm.isExpire() {
			return "", ErrNil
		}

		if itm.value == nil {
			return "", ErrNil
		}
		switch v := itm.value.(type) {
		case string:
			return v, nil
		case *string:
			return *v, nil
		default:
			return fmt.Sprint(itm.value), nil
		}
	}
	return "", ErrNil
}
func (c *memoryCache) GetMulti(keys []string) ([]string, error) {
	var rs []string
	var err error
	for _, key := range keys {
		v, err1 := c.Get(key)
		if err1 != nil && err1 != ErrNil {
			err = err1
		}
		rs = append(rs, v)
	}
	return rs, err
}
func (c *memoryCache) PutObject(key string, val interface{}, expire ...time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.put(key, val, expire...)
	return nil
}

//valptr - object ptr
func (c *memoryCache) GetObject(key string, valptr interface{}) error {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if itm, ok := c.items[key]; ok {
		if itm.isExpire() {
			return ErrNil
		}
		v := reflect.ValueOf(itm.value)
		if v.Type().Kind() == reflect.Ptr {
			reflect.ValueOf(valptr).Elem().Set(v.Elem())
		} else {
			reflect.ValueOf(valptr).Elem().Set(v)
		}
		return nil
	}
	return ErrNil
}
func (c *memoryCache) Delete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.items[key]; !ok {
		//return errors.New("key not exist")
		return nil
	}
	delete(c.items, key)
	if _, ok := c.items[key]; ok {
		return errors.New("delete key error")
	}
	return nil
}
func (c *memoryCache) Incr(key string) error {
	c.lock.RLock()
	defer c.lock.RUnlock()
	itm, ok := c.items[key]
	if !ok {
		itm = c.put(key, 0, 0)
	} else {
		if itm.value == nil {
			itm.value = 0
		}
	}
	switch itm.value.(type) {
	case int:
		itm.value = itm.value.(int) + 1
	case int32:
		itm.value = itm.value.(int32) + 1
	case int64:
		itm.value = itm.value.(int64) + 1
	case uint:
		itm.value = itm.value.(uint) + 1
	case uint32:
		itm.value = itm.value.(uint32) + 1
	case uint64:
		itm.value = itm.value.(uint64) + 1
	default:
		return errors.New("item val is not (u)int (u)int32 (u)int64")
	}
	return nil
}
func (c *memoryCache) Decr(key string) error {
	c.lock.RLock()
	defer c.lock.RUnlock()
	itm, ok := c.items[key]
	if !ok {
		itm = c.put(key, 0, 0)
	} else {
		if itm.value == nil {
			itm.value = 0
		}
	}

	switch itm.value.(type) {
	case int:
		itm.value = itm.value.(int) - 1
	case int64:
		itm.value = itm.value.(int64) - 1
	case int32:
		itm.value = itm.value.(int32) - 1
	case uint:
		if itm.value.(uint) > 0 {
			itm.value = itm.value.(uint) - 1
		} else {
			return errors.New("item val is less than 0")
		}
	case uint32:
		if itm.value.(uint32) > 0 {
			itm.value = itm.value.(uint32) - 1
		} else {
			return errors.New("item val is less than 0")
		}
	case uint64:
		if itm.value.(uint64) > 0 {
			itm.value = itm.value.(uint64) - 1
		} else {
			return errors.New("item val is less than 0")
		}
	default:
		return errors.New("item val is not int int64 int32")
	}
	return nil
}
func (c *memoryCache) Exists(key string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if itm, ok := c.items[key]; ok {
		if itm.isExpire() {
			return false
		}
		return true
	}
	return false
}

func (c *memoryCache) setExpire(key string, expire time.Duration) *memoryEntry {
	var me *memoryEntry
	if itm, ok := c.items[key]; ok {
		itm.createdtime = time.Now()
		itm.expire = expire
		me = itm
	} else {
		me = &memoryEntry{value: nil, createdtime: time.Now(), expire: expire}
		c.items[key] = me
	}
	return me
}

func (c *memoryCache) SetExpire(key string, expire ...time.Duration) error {
	if len(expire) == 0 {
		return nil
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.setExpire(key, expire[0])
	return nil
}
func (c *memoryCache) NewMap(name string, expire ...time.Duration) (Map, error) {
	var timeout time.Duration
	if len(expire) > 0 {
		timeout = expire[0]
	} else if c.defaultExpire > 0 {
		timeout = c.defaultExpire
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	itm := c.setExpire(name, timeout)
	if itm.value == nil {
		itm.value = newMemoryMap(c, name)
	} else if _, ok := itm.value.(Map); !ok {
		itm.value = newMemoryMap(c, name)
	}
	return itm.value.(Map), nil
}

//type memoryMap map[string]interface{}
type memoryMap struct {
	c    *memoryCache
	name string
	data map[string]interface{}
	lock *sync.RWMutex
}

func newMemoryMap(c *memoryCache, name string) *memoryMap {
	return &memoryMap{c, name, make(map[string]interface{}), new(sync.RWMutex)}
}

func (m *memoryMap) Put(key string, val string) error {
	if m.c.lget(m.name) == nil {
		return errors.New("cache: map(" + m.name + ")." + key + " is expired")
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[key] = val
	return nil
}
func (m *memoryMap) Get(key string) (string, error) {
	if m.c.lget(m.name) == nil {
		return "", ErrNil
	}
	m.lock.RLock()
	defer m.lock.RUnlock()

	value := m.data[key]
	if value == nil {
		return "", ErrNil
	}
	switch v := value.(type) {
	case string:
		return v, nil
	case *string:
		return *v, nil
	default:
		return fmt.Sprint(value), nil
	}

}
func (m *memoryMap) GetMulti(keys []string) ([]string, error) {
	var rs []string
	var err error
	if m.c.lget(m.name) == nil {
		for _, _ = range keys {
			rs = append(rs, "")
		}
		err = ErrNil
	} else {
		for _, key := range keys {
			v, err1 := m.Get(key)
			if err1 != nil && err1 != ErrNil {
				err = err1
			}
			rs = append(rs, v)
		}
	}
	return rs, err
}

func (m *memoryMap) PutObject(key string, val interface{}) error {
	if m.c.lget(m.name) == nil {
		return errors.New("cache: map(" + m.name + ")." + key + " is expired")
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[key] = val
	return nil
}

func (m *memoryMap) GetObject(key string, valptr interface{}) error {
	if m.c.lget(m.name) == nil {
		return ErrNil
	}
	m.lock.RLock()
	defer m.lock.RUnlock()

	if value, ok := m.data[key]; ok {
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		reflect.ValueOf(valptr).Elem().Set(v)
		return nil
	}
	return ErrNil

}

func (m *memoryMap) Delete(key string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.data[key]; !ok {
		//return errors.New("key not exist")
		return nil
	}
	delete(m.data, key)
	if _, ok := m.data[key]; ok {
		return errors.New("delete key error")
	}
	return nil
}

func (m *memoryMap) Incr(key string) error {
	if m.c.lget(m.name) == nil {
		return errors.New("cache: map(" + m.name + ")." + key + " is expired")
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	value, ok := m.data[key]
	if !ok {
		value = 0
	} else {
		if value == nil {
			value = 0
		}
	}

	switch v := value.(type) {
	case int:
		value = v + 1
	case int32:
		value = v + 1
	case int64:
		value = v + 1
	case uint:
		value = v + 1
	case uint32:
		value = v + 1
	case uint64:
		value = v + 1
	default:
		return errors.New("item val is not (u)int (u)int32 (u)int64")
	}
	m.data[key] = value
	return nil
}

func (m *memoryMap) Decr(key string) error {
	if m.c.lget(m.name) == nil {
		return errors.New("cache: map(" + m.name + ")." + key + " is expired")
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	value, ok := m.data[key]
	if !ok {
		value = 0
	} else {
		if value == nil {
			value = 0
		}
	}

	switch v := value.(type) {
	case int:
		value = v - 1
	case int64:
		value = v - 1
	case int32:
		value = v - 1
	case uint:
		value = v - 1
	case uint32:
		value = v - 1
	case uint64:
		value = v - 1
	default:
		return errors.New("item val is not int int64 int32")
	}
	m.data[key] = value
	return nil
}

func (m *memoryMap) Exists(key string) bool {
	if m.c.lget(m.name) == nil {
		return false
	}
	m.lock.RLock()
	defer m.lock.RUnlock()

	if _, ok := m.data[key]; ok {
		return true
	}
	return false
}

func (m *memoryMap) Size() (int, error) {
	if m.c.lget(m.name) == nil {
		return 0, errors.New("cache: map(" + m.name + ") is expired")
	}
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.data), nil
}

func (m *memoryMap) Clear() error {
	if m.c.lget(m.name) == nil {
		return errors.New("cache: map(" + m.name + ") is expired")
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data = make(map[string]interface{})
	return nil
}

func init() {
	Register("memory", NewMemoryCache)
}

//-------------------------------------------------------------------------------
//type MemoryCache struct {
//	maxEntries int
//	ll         *list.List
//	cache      map[string]*list.Element
//	duration   time.Duration
//	lock       sync.RWMutex
//}

//type entry struct {
//	key     string
//	value   interface{}
//	timeout time.Time
//}

//func New(maxEntries int, duration time.Duration) *MemoryCache {
//	var c = &MemoryCache{
//		maxEntries: maxEntries,
//		ll:         list.New(),
//		cache:      make(map[string]*list.Element),
//		duration:   duration,
//	}
//	if duration > 0 {
//		go func() {
//			for {
//				var d = c.check()
//				time.Sleep(d)
//			}
//		}()
//	}
//	return c
//}

//func (c *MemoryCache) check() time.Duration {
//	var now = time.Now()
//	c.lock.Lock()
//	defer c.lock.Unlock()
//	for {
//		var ele = c.ll.Back()
//		if ele == nil {
//			return c.duration + 10*time.Second
//		}
//		var timeout = ele.Value.(*entry).timeout
//		if now.Before(timeout) {
//			var dur = timeout.Sub(now)
//			if dur < 10*time.Second {
//				dur = 10 * time.Second
//			}
//			return dur
//		}
//		c.removeElement(ele)
//	}
//}

//func (c *MemoryCache) Put(key string, value interface{}, expire ...time.Duration) {
//	c.lock.Lock()
//	defer c.lock.Unlock()
//	c.put(key, value)
//}

//func (c *MemoryCache) put(key string, value interface{}, expire ...time.Duration) {

//	timeout := func() time.Time {
//		if len(expire) > 0 {
//			return time.Now().Add(expire[0])
//		} else {
//			return time.Now().Add(c.duration)
//		}
//	}

//	if ee, ok := c.cache[key]; ok {
//		c.ll.MoveToFront(ee)
//		var v = ee.Value.(*entry)
//		v.value = value
//		v.timeout = timeout() //time.Now().Add(c.duration)
//		return
//	}
//	//var ele = c.ll.PushFront(&entry{key, value, time.Now().Add(c.duration)})
//	var ele = c.ll.PushFront(&entry{key, value, timeout()})
//	c.cache[key] = ele

//	if c.maxEntries == 0 {
//		return
//	}
//	for c.ll.Len() > c.maxEntries {
//		c.removeOldest()
//	}
//}

//func (c *MemoryCache) isExpire(ele *list.Element) bool {
//	return time.Now().After(ele.Value.(*entry).timeout)
//}

//func (c *MemoryCache) get(key string) interface{} {
//	if ele, hit := c.cache[key]; hit && !c.isExpire(ele) {
//		c.ll.MoveToFront(ele)
//		return ele.Value.(*entry).value
//	}
//	return nil
//}

////func (c *MemoryCache) GetIfPresent(key string) interface{} {
////	c.lock.Lock()
////	defer c.lock.Unlock()

////	return c.get(key)
////}

//func (c *MemoryCache) Exists(key string) bool {
//	c.lock.RLock()
//	defer c.lock.RUnlock()
//	_, ok := c.cache[key]
//	return ok
//}

//func (c *MemoryCache) Get(key string, setter ...func() interface{}) interface{} {
//	c.lock.Lock()
//	defer c.lock.Unlock()

//	var v = c.get(key)
//	if v != nil {
//		return v
//	}

//	if len(setter) > 0 {
//		v = setter[0]()
//		if v != nil {
//			c.put(key, v)
//		}
//	}
//	return v
//}

//func (c *MemoryCache) Remove(key string) {
//	c.lock.Lock()
//	defer c.lock.Unlock()

//	if ele, hit := c.cache[key]; hit {
//		c.removeElement(ele)
//	}
//}

//func (c *MemoryCache) removeOldest() {
//	var ele = c.ll.Back()
//	if ele != nil {
//		c.removeElement(ele)
//	}
//}

//func (c *MemoryCache) removeElement(e *list.Element) {
//	c.ll.Remove(e)
//	var kv = e.Value.(*entry)
//	delete(c.cache, kv.key)
//}

//func (c *MemoryCache) Len() int {
//	c.lock.Lock()
//	defer c.lock.Unlock()

//	return c.ll.Len()
//}
