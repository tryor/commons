package cache

import (
	"errors"
	"time"
)

//multi-level cache

func NewLNCache(cs ...Cache) Cache {
	switch len(cs) {
	case 0:
		return nil
	case 1:
		return cs[0]
	case 2:
		return NewL2Cache(cs[0], cs[1])
	default:
		i := len(cs) - 1
		c := NewL2Cache(cs[i-1], cs[i])
		return NewLNCache(append(cs[0:i-1], c)...)
	}
}

func NewL3Cache(c1 Cache, c2 Cache, c3 Cache) Cache {
	return NewL2Cache(c1, NewL2Cache(c2, c3))
}

func NewL2Cache(c1 Cache, c2 Cache) Cache {
	return &l2Cache{c1: c1, c2: c2}
}

type l2Cache struct {
	c1 Cache
	c2 Cache
}

func (lc *l2Cache) Error(err1, err2 error) error {
	if err1 == nil {
		return err2
	}
	if err2 == nil {
		return err1
	}
	return errors.New(err1.Error() + ", " + err2.Error())
}

func (lc *l2Cache) Put(key string, val string, expire ...time.Duration) error {
	var err1, err2 error
	switch len(expire) {
	case 0:
		err1 = lc.c1.Put(key, val)
		err2 = lc.c2.Put(key, val)
	case 1:
		err1 = lc.c1.Put(key, val, expire[0])
		err2 = lc.c2.Put(key, val)
	default:
		err1 = lc.c1.Put(key, val, expire[0])
		err2 = lc.c2.Put(key, val, expire[1])
	}
	return lc.Error(err1, err2)
}

func (lc *l2Cache) Get(key string) (string, error) {

	v, err := lc.c1.Get(key)
	if err == nil {
		return v, nil
	}

	v, err = lc.c2.Get(key)
	if err == nil {
		lc.c1.Put(key, v)
		return v, nil
	}

	return "", err

}

func (lc *l2Cache) GetMulti(keys []string) ([]string, error) {
	vs, err := lc.c1.GetMulti(keys)
	if err == nil {
		nullNums := 0
		for i := 0; i < len(vs); i++ {
			if vs[i] == "" {
				nullNums++
			}
		}
		if nullNums < len(vs) {
			return vs, nil
		}
	}
	vs, err = lc.c2.GetMulti(keys)
	if err == nil {
		for i, k := range keys {
			lc.c1.Put(k, vs[i])
		}
		return vs, nil
	}
	return nil, err
}

func (lc *l2Cache) PutObject(key string, val interface{}, expire ...time.Duration) error {
	var err1, err2 error
	switch len(expire) {
	case 0:
		err1 = lc.c1.PutObject(key, val)
		err2 = lc.c2.PutObject(key, val)
	case 1:
		err1 = lc.c1.PutObject(key, val, expire[0])
		err2 = lc.c2.PutObject(key, val)
	default:
		err1 = lc.c1.PutObject(key, val, expire[0])
		err2 = lc.c2.PutObject(key, val, expire[1])
	}
	return lc.Error(err1, err2)
}

func (lc *l2Cache) GetObject(key string, valptr interface{}) error {

	err := lc.c1.GetObject(key, valptr)
	if err == nil {
		return nil
	}

	err = lc.c2.GetObject(key, valptr)
	if err == nil {
		lc.c1.PutObject(key, valptr)
		return nil
	}

	return err
}

func (lc *l2Cache) Delete(key string) error {
	err1 := lc.c1.Delete(key)
	err2 := lc.c2.Delete(key)
	return lc.Error(err1, err2)
}
func (lc *l2Cache) Incr(key string) error {
	err1 := lc.c1.Incr(key)
	err2 := lc.c2.Incr(key)
	return lc.Error(err1, err2)
}
func (lc *l2Cache) Decr(key string) error {
	err1 := lc.c1.Decr(key)
	err2 := lc.c2.Decr(key)
	return lc.Error(err1, err2)
}
func (lc *l2Cache) Exists(key string) bool {
	if lc.c1.Exists(key) {
		return true
	}
	return lc.c2.Exists(key)
}

func (lc *l2Cache) SetExpire(key string, expire ...time.Duration) error {

	var err1, err2 error
	switch len(expire) {
	case 0:
	case 1:
		err1 = lc.c1.SetExpire(key, expire[0])
	default:
		err1 = lc.c1.SetExpire(key, expire[0])
		err2 = lc.c2.SetExpire(key, expire[1:]...)
	}
	return lc.Error(err1, err2)
}

func (lc *l2Cache) NewMap(name string, expire ...time.Duration) (Map, error) {
	m := &l2CacheMap{}
	var err1, err2 error
	switch len(expire) {
	case 0:
		m.m1, err1 = lc.c1.NewMap(name)
		m.m2, err2 = lc.c2.NewMap(name)
	case 1:
		m.m1, err1 = lc.c1.NewMap(name, expire[0])
		m.m2, err2 = lc.c2.NewMap(name)
	default:
		m.m1, err1 = lc.c1.NewMap(name, expire[0])
		m.m2, err2 = lc.c2.NewMap(name, expire[1:]...)
	}
	return m, lc.Error(err1, err2)
}

type l2CacheMap struct {
	lc *l2Cache
	m1 Map
	m2 Map
}

func (m *l2CacheMap) Put(key string, val string) error {
	err1 := m.m1.Put(key, val)
	err2 := m.m2.Put(key, val)
	return m.lc.Error(err1, err2)
}

func (m *l2CacheMap) Get(key string) (string, error) {
	v, err := m.m1.Get(key)
	if err == nil {
		return v, nil
	}

	v, err = m.m2.Get(key)
	if err == nil {
		m.m1.Put(key, v)
		return v, nil
	}

	return "", err
}
func (m *l2CacheMap) GetMulti(keys []string) ([]string, error) {
	vs, err := m.m1.GetMulti(keys)
	if err == nil {
		nullNums := 0
		for i := 0; i < len(vs); i++ {
			if vs[i] == "" {
				nullNums++
			}
		}
		if nullNums < len(vs) {
			return vs, nil
		}
	}
	vs, err = m.m2.GetMulti(keys)
	if err == nil {
		for i, k := range keys {
			m.m1.Put(k, vs[i])
		}
		return vs, nil
	}
	return nil, err
}

func (m *l2CacheMap) PutObject(key string, val interface{}) error {
	err1 := m.m1.PutObject(key, val)
	err2 := m.m2.PutObject(key, val)
	return m.lc.Error(err1, err2)
}

func (m *l2CacheMap) GetObject(key string, valptr interface{}) error {
	err := m.m1.GetObject(key, valptr)
	if err == nil {
		return nil
	}

	err = m.m2.GetObject(key, valptr)
	if err == nil {
		m.m1.PutObject(key, valptr)
		return nil
	}

	return err
}

func (m *l2CacheMap) Delete(key string) error {
	err1 := m.m1.Delete(key)
	err2 := m.m2.Delete(key)
	return m.lc.Error(err1, err2)
}
func (m *l2CacheMap) Incr(key string) error {
	err1 := m.m1.Incr(key)
	err2 := m.m2.Incr(key)
	return m.lc.Error(err1, err2)
}
func (m *l2CacheMap) Decr(key string) error {
	err1 := m.m1.Decr(key)
	err2 := m.m2.Decr(key)
	return m.lc.Error(err1, err2)
}
func (m *l2CacheMap) Exists(key string) bool {
	if m.m1.Exists(key) {
		return true
	}
	return m.m2.Exists(key)
}
func (m *l2CacheMap) Size() (int, error) {
	n, err := m.m1.Size()
	if err == nil {
		return n, nil
	}
	return m.m2.Size()
}
func (m *l2CacheMap) Clear() error {
	err1 := m.m1.Clear()
	err2 := m.m2.Clear()
	return m.lc.Error(err1, err2)
}
