#安装
```
go get -u github.com/trygo/util
```

#使用
```
package main

import (
    "github.com/trygo/util/cache"
    "log"
    "time"
)

func main() {
    var c = cache.New(100, time.Second*2)
    c.Put("hello", "world")
    log.Println(c.GetIfPresent("hello"))
    time.Sleep(time.Second)
    log.Println(c.GetIfPresent("hello"))
    time.Sleep(time.Second * 20)
    log.Println(c.GetIfPresent("hello"))
}
```
输出：
```
-> % go run a.go
2014/10/01 08:00:00 world
2014/10/01 08:00:01 world
2014/10/01 08:00:21 <nil>
```
注意：
出于性能考虑，检查频率为10秒



