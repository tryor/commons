package cache

import (
	"fmt"
	"testing"
	"time"
)

type V struct {
	V1 string
	V2 int
	V3 float32
}

func Test_NewLNCache(t *testing.T) {
	config := `{"gccyc":1, "defaultExpire":2}`
	memory1, err := NewCache("memory", config)
	if err != nil {
		t.Fatal(err)
	}

	config = `{"gccyc":1, "defaultExpire":5}`
	memory2, err := NewCache("memory", config)
	if err != nil {
		t.Fatal(err)
	}

	config = `{"gccyc":1, "defaultExpire":8}`
	memory3, err := NewCache("memory", config)
	if err != nil {
		t.Fatal(err)
	}

	config = `{"gccyc":1, "defaultExpire":10}`
	memory4, err := NewCache("memory", config)
	if err != nil {
		t.Fatal(err)
	}

	config = `{"addr":"127.0.0.1:6379", "password":"", "defaultExpire":"15"}`
	redis, err := NewCache("redis", config)
	if err != nil {
		t.Fatal(err)
	}

	cache := NewLNCache(memory1, memory2, memory3, memory4, redis)

	strv := `v1`
	err = cache.Put("ln_k1", strv)
	if err != nil {
		t.Fatal(err)
	}

	v, err := cache.Get("ln_k1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
	if v != strv {
		t.Errorf("%s not equal %s", v, strv)
	}

	//time.Sleep(time.Second * 3)
	v, err = memory1.Get("ln_k1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("v:", v)

	if !memory1.Exists("ln_k1") {
		t.Fatal("memory1: ln_k1 not is exist!")
	}

	if !memory2.Exists("ln_k1") {
		t.Fatal("memory2: ln_k1 not is exist!")
	}

	if !memory3.Exists("ln_k1") {
		t.Fatal("memory3: ln_k1 not is exist!")
	}

	if !memory4.Exists("ln_k1") {
		t.Fatal("memory4: ln_k1 not is exist!")
	}

	if !redis.Exists("ln_k1") {
		t.Fatal("redis: ln_k1 not is exist!")
	}

	time.Sleep(time.Second * 3)

	if memory1.Exists("ln_k1") {
		t.Fatal("memory1: ln_k1 is exist!")
	}

	if !memory2.Exists("ln_k1") {
		t.Fatal("memory2: ln_k1 not is exist!")
	}

	if !memory3.Exists("ln_k1") {
		t.Fatal("memory3: ln_k1 not is exist!")
	}

	if !memory4.Exists("ln_k1") {
		t.Fatal("memory4: ln_k1 not is exist!")
	}

	time.Sleep(time.Second * 8)

	if memory4.Exists("ln_k1") {
		t.Fatal("memory4: ln_k1 is exist!")
	}

	if !redis.Exists("ln_k1") {
		t.Fatal("redis: ln_k1 not is exist!")
	}

	v, err = cache.Get("ln_k1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("v:", v)

	if !memory1.Exists("ln_k1") {
		t.Fatal("memory1: ln_k1 not is exist!")
	}

}

func Test_L2Cache_Other(t *testing.T) {
	config := `{"gccyc":1, "defaultExpire":3}`
	memory, err := NewCache("memory", config)
	if err != nil {
		t.Fatal(err)
	}

	config = `{"addr":"127.0.0.1:6379", "password":"", "defaultExpire":"5"}`
	redis, err := NewCache("redis", config)
	if err != nil {
		t.Fatal(err)
	}

	cache := NewL2Cache(memory, redis)

	err = cache.Incr("l2Incr")
	if err != nil {
		t.Fatal(err)
	}
	cache.SetExpire("l2Incr", time.Second*3, time.Second*5)

	var l2Incr int
	err = cache.GetObject("l2Incr", &l2Incr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("l2Incr:", l2Incr)

}

func Test_L2Cache_PutGetObject(t *testing.T) {
	config := `{"gccyc":1, "defaultExpire":3}`
	memory, err := NewCache("memory", config)
	if err != nil {
		t.Fatal(err)
	}

	config = `{"addr":"127.0.0.1:6379", "password":"", "defaultExpire":"600"}`
	redis, err := NewCache("redis", config)
	if err != nil {
		t.Fatal(err)
	}

	cache := NewL2Cache(memory, redis)

	v := V{"v1", 123, 456.789}
	err = cache.PutObject("ok1", &v, time.Second*2, time.Second*4)
	if err != nil {
		t.Fatal(err)
	}

	var rv V
	err = cache.GetObject("ok1", &rv)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rv)
	if rv != v {
		t.Fatal("rv != v")
	}

	var mrv V
	err = memory.GetObject("ok1", &mrv)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("memory, ", mrv)
	if mrv != v {
		t.Fatal("mrv != v")
	}

	var rrv V
	err = redis.GetObject("ok1", &rrv)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("redis, ", rrv)
	if rrv != v {
		t.Fatal("rrv != v")
	}

}

func Test_L2Cache(t *testing.T) {
	config := `{"gccyc":1, "defaultExpire":3}`
	memory, err := NewCache("memory", config)
	if err != nil {
		t.Fatal(err)
	}

	config = `{"addr":"127.0.0.1:6379", "password":"", "defaultExpire":"600"}`
	redis, err := NewCache("redis", config)
	if err != nil {
		t.Fatal(err)
	}

	cache := NewL2Cache(memory, redis)

	testL2Cache(t, cache, memory, redis)

	testMap(t, cache, "l2")
}

func testL2Cache(t *testing.T, cache Cache, memory Cache, redis Cache) {
	strv := `test cache string 1 ....`
	err := cache.Put("k1", strv, time.Second*2, time.Second*4)
	if err != nil {
		t.Fatal(err)
	}

	v, err := cache.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
	if v != strv {
		t.Errorf("%s not equal %s", v, strv)
	}

	v, err = memory.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("memory:", v)
	if v != strv {
		t.Errorf("%s not equal %s", v, strv)
	}

	v, err = redis.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("redis:", v)
	if v != strv {
		t.Errorf("%s not equal %s", v, strv)
	}

	time.Sleep(time.Second * 3)

	v, err = memory.Get("k1")
	if err == nil || err != nil && err != ErrNil {
		t.Fatal(err)
	}
	t.Log("memory, err:", err)

	v, err = redis.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("redis:", v)
	if v != strv {
		t.Errorf("%s not equal %s", v, strv)
	}

	v, err = cache.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
	if v != strv {
		t.Errorf("%s not equal %s", v, strv)
	}

	v, err = memory.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("2 memory, err:%v, v:%v", err, v)

	time.Sleep(time.Second * 2)

	v, err = redis.Get("k1")
	if err == nil || err != nil && err != ErrNil {
		t.Fatal(err)
	}
	t.Log("redis, err:", err)

	time.Sleep(time.Second * 2)

	v, err = memory.Get("k1")
	if err == nil || err != nil && err != ErrNil {
		t.Fatal(err)
	}
	t.Logf("3 memory, err:%v, v:%v", err, v)

	err = cache.Put("k2", "v2")
	if err != nil {
		t.Fatal(err)
	}

	err = cache.Put("k3", "v3")
	if err != nil {
		t.Fatal(err)
	}

	vs, err := cache.GetMulti([]string{"k2", "k3", "k4"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(vs)
	if len(vs) != 3 {
		t.Fatal("len(vs) != 3")
	}

	err = cache.Delete("k2")
	if err != nil {
		t.Fatal(err)
	}
	if cache.Exists("k2") {
		t.Fatal("k2 is exist!")
	}

	err = cache.SetExpire("k3", time.Second*1, time.Second*3)
	if err != nil {
		t.Fatal(err)
	}
	if !cache.Exists("k3") {
		t.Fatal("k3 not is exist!")
	}

	time.Sleep(time.Second * 2)
	if !cache.Exists("k3") {
		t.Fatal("k3 not is exist!")
	}

	if memory.Exists("k3") {
		t.Fatal("k3 is exist!")
	}

	if !redis.Exists("k3") {
		t.Fatal("k3 not is exist!")
	}

	time.Sleep(time.Second * 2)
	if cache.Exists("k3") {
		t.Fatal("k3 is exist!")
	}

}

func Test_Memory(t *testing.T) {
	config := `{"gccyc":1}`
	cache, err := NewCache("memory", config)
	if err != nil {
		t.Fatal(err)
	}
	testCache(t, cache)
	testMap(t, cache, "m")
}

func Test_Redis(t *testing.T) {
	config := `{"addr":"127.0.0.1:6379", "password":"", "dbNum":"0", "maxIdleConns":"9", "connIdleTimeout":"310", "noTesttime":"61", "defaultExpire":"600"}`
	cache, err := NewCache("redis", config)
	if err != nil {
		t.Fatal(err)
	}
	testCache(t, cache)
	testMap(t, cache, "r")
}

func testMap(t *testing.T, cache Cache, p string) {
	m, err := cache.NewMap(p+"_map001", time.Second*4, time.Second*6)
	if err != nil {
		t.Fatal(err)
	}
	mv1 := "test map ..."
	m.Put("mv1", mv1)
	rmv1, err := m.Get("mv1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(mv1)
	if mv1 != rmv1 {
		t.Fatalf("%s != %s", mv1, rmv1)
	}
	objv := V{"test map ...", 678, 890.455}
	err = m.PutObject("mobj1", &objv)
	if err != nil {
		t.Fatal(err)
	}
	var robjv V
	m.GetObject("mobj1", &robjv)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("[map] robjv:", robjv)

	err = m.Incr("mincr")
	if err != nil {
		t.Fatal(err)
	}

	var incr int
	err = m.GetObject("mincr", &incr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("[map] incr:", incr)

	err = m.Decr("mdecr")
	if err != nil {
		t.Fatal(err)
	}

	var decr int
	err = m.GetObject("mdecr", &decr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("[map] decr:", decr)

	size, err := m.Size()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("[map], m.Size():", size)
	if size != 4 {
		t.Fatalf("size != 4")
	}

	mvs, err := m.GetMulti([]string{"mv1", "mobj1", "mincr", "mdecr", "other"})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("[map], m.GetMulti()(%v):%v", len(mvs), mvs)
	if len(mvs) != 5 {
		t.Fatalf("len(mvs) != 5")
	}

	if !m.Exists("mv1") {
		t.Fatalf("mv1 not is exist")
	}

	err = m.Delete("mv1")
	if err != nil {
		t.Fatal(err)
	}

	if m.Exists("mv1") {
		t.Fatalf("mv1 is exist")
	}

	size, err = m.Size()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("[map], m.Size():", size)
	if size != 3 {
		t.Fatalf("size != 3")
	}

	err = m.Clear()
	if err != nil {
		t.Fatal(err)
	}

	size, err = m.Size()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("[map], m.Size():", size)
	if size != 0 {
		t.Fatalf("size != 0")
	}

	m2, err := cache.NewMap(p+"_map002", time.Second*1, time.Second*2)
	if err != nil {
		t.Fatal(err)
	}

	err = m2.Incr("m2incr")
	if err != nil {
		t.Fatal(err)
	}

	err = m2.Put("mk5", "mv5")
	if err != nil {
		t.Fatal(err)
	}
	mv5, err := m2.Get("mk5")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("[map], mv5:", mv5)
	if mv5 != "mv5" {
		t.Fatal("mk5 != mv5")
	}

	time.Sleep(time.Second * 3)
	mv5, err = m2.Get("mk5")
	if err != ErrNil {
		t.Fatalf("err:%v, mv5:%v", err, mv5)
	}
	t.Log("[map], mv5:", mv5)
	if mv5 != "" {
		t.Fatal("mk5 != \"\"")
	}

}

func testCache(t *testing.T, cache Cache) {

	strv := `test cache string 1 ....`
	err := cache.Put("k1", strv, time.Second*10)
	if err != nil {
		t.Fatal(err)
	}

	v, err := cache.Get("k1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
	if v != strv {
		t.Errorf("%s not equal %s", v, strv)
	}

	objv := V{"test...", 123, 123.455}
	err = cache.PutObject("objk1", objv, time.Second*10)
	if err != nil {
		t.Fatal(err)
	}

	var robjv V
	err = cache.GetObject("objk1", &robjv)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("robjv:", robjv)

	if robjv != objv {
		t.Errorf("%v not equal %v", objv, robjv)
	}

	strv2 := `test cache string 2 ....`
	err = cache.Put("k2", strv2, time.Second*10)
	if err != nil {
		t.Fatal(err)
	}

	vs, err := cache.GetMulti([]string{"k0", "k1", "k11", "k2", "k12"})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("vs(%d):%v", len(vs), vs)

	if !cache.Exists("k1") {
		t.Fatal("k1 not exist!")
	}

	err = cache.Delete("k1")
	if err != nil {
		t.Fatal(err)
	}

	if cache.Exists("k1") {
		t.Fatal("k1 is exist!")
	}

	err = cache.Incr("incr")
	if err != nil {
		t.Fatal(err)
	}
	cache.SetExpire("incr", time.Second*5)

	var incr int
	err = cache.GetObject("incr", &incr)
	if err != nil || incr <= 0 {
		t.Fatal(err)
	}
	t.Log("incr:", incr)

	err = cache.Decr("decr")
	if err != nil {
		t.Fatal(err)
	}
	cache.SetExpire("decr", time.Second*5)

	var decr int
	err = cache.GetObject("decr", &decr)
	if err != nil || decr > 0 {
		t.Fatal(err)
	}
	t.Log("decr:", decr)

	err = cache.Put("k5", "v5", time.Second*1)
	if err != nil {
		t.Fatal(err)
	}
	v5, err := cache.Get("k5")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("v5:", v5)
	if v5 != "v5" {
		t.Fatal("k5 != v5")
	}
	time.Sleep(time.Second * 2)
	v5, err = cache.Get("k5")
	if err != ErrNil {
		t.Fatalf("err:%v, v5:%v", err, v5)
	}
	t.Log("v5:", v5)
	if v5 != "" {
		t.Fatal("k5 != \"\"")
	}

}
