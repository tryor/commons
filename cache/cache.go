package cache

import (
	"errors"
	"fmt"
	"time"
)

type Cache interface {
	Put(key string, val string, expire ...time.Duration) error
	Get(key string) (string, error)
	GetMulti(keys []string) ([]string, error)

	PutObject(key string, val interface{}, expire ...time.Duration) error
	//valptr - object ptr
	GetObject(key string, valptr interface{}) error

	Delete(key string) error
	Incr(key string) error
	Decr(key string) error
	Exists(key string) bool
	SetExpire(key string, expire ...time.Duration) error
	NewMap(name string, expire ...time.Duration) (Map, error)
}

type Map interface {
	Put(key string, val string) error
	Get(key string) (string, error)
	GetMulti(keys []string) ([]string, error)

	PutObject(key string, val interface{}) error
	GetObject(key string, valptr interface{}) error

	Delete(key string) error
	Incr(key string) error
	Decr(key string) error
	Exists(key string) bool
	Size() (int, error)
	Clear() error
}

type CacheInitializer interface {
	Init(config string) error
}

var adapters = make(map[string]func() Cache)

func Register(name string, adapter func() Cache) {
	if adapter == nil {
		panic("cache: register adapter is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("cache: register called twice for adapter " + name)
	}

	adapters[name] = adapter
}

func NewCache(adapterName, config string) (adapter Cache, err error) {
	instanceFunc, ok := adapters[adapterName]
	if !ok {
		err = fmt.Errorf("cache: unknown adapter name %q", adapterName)
		return
	}
	adapter = instanceFunc()

	if initor, ok := adapter.(CacheInitializer); ok {
		err = initor.Init(config)
		if err != nil {
			adapter = nil
		}
	} else {
		err = errors.New("Init(config string) not implemented, adapter is " + adapterName)
	}
	return
}

var (
	ErrNil = errors.New("cache: nil returned")
)
