package cache

import (
	"container/list"
	"sync"
	"time"
)

type Cache struct {
	maxEntries int
	ll         *list.List
	cache      map[string]*list.Element
	duration   time.Duration
	lock       sync.Mutex
}

type entry struct {
	key     string
	value   interface{}
	timeout time.Time
}

func New(maxEntries int, duration time.Duration) *Cache {
	var c = &Cache{
		maxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[string]*list.Element),
		duration:   duration,
	}
	if duration > 0 {
		go func() {
			for {
				var d = c.check()
				time.Sleep(d)
			}
		}()
	}
	return c
}

func (c *Cache) check() time.Duration {
	var now = time.Now()
	c.lock.Lock()
	defer c.lock.Unlock()
	for {
		var ele = c.ll.Back()
		if ele == nil {
			return c.duration + 10*time.Second
		}
		var timeout = ele.Value.(*entry).timeout
		if now.Before(timeout) {
			var dur = timeout.Sub(now)
			if dur < 10*time.Second {
				dur = 10 * time.Second
			}
			return dur
		}
		c.removeElement(ele)
	}
}

func (c *Cache) Put(key string, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.put(key, value)
}

func (c *Cache) put(key string, value interface{}) {
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		var v = ee.Value.(*entry)
		v.value = value
		v.timeout = time.Now().Add(c.duration)
		return
	}
	var ele = c.ll.PushFront(&entry{key, value, time.Now().Add(c.duration)})
	c.cache[key] = ele

	if c.maxEntries == 0 {
		return
	}
	for c.ll.Len() > c.maxEntries {
		c.removeOldest()
	}
}

func (c *Cache) get(key string) interface{} {
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value
	}
	return nil
}

func (c *Cache) GetIfPresent(key string) interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.get(key)
}

func (c *Cache) Get(key string, setter func() interface{}) interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()

	var v = c.get(key)
	if v != nil {
		return v
	}
	v = setter()
	if v != nil {
		c.put(key, v)
	}
	return v
}

func (c *Cache) Remove(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

func (c *Cache) removeOldest() {
	var ele = c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	var kv = e.Value.(*entry)
	delete(c.cache, kv.key)
}

func (c *Cache) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.ll.Len()
}
