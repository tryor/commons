package util

import (
	//"bytes"
	"crypto/md5"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	//"unicode"
)

func ParseUint(s string, def ...uint64) uint64 {
	if s == "" {
		if len(def) > 0 {
			return def[0]
		} else {
			return 0
		}
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func ParseInt(s string, def ...int64) int64 {
	if s == "" {
		if len(def) > 0 {
			return def[0]
		} else {
			return 0
		}
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func Atoi(s string, def ...int) int {
	if s == "" {
		if len(def) > 0 {
			return def[0]
		} else {
			return 0
		}
	}
	v, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		panic(err)
	}
	return int(v)
}

func Md5(s string) string {
	inst := md5.New()
	inst.Write([]byte(s))
	return fmt.Sprintf("%x", inst.Sum([]byte("")))
}

func WebTime(t time.Time) string {
	ftime := t.Format(time.RFC1123)
	if strings.HasSuffix(ftime, "UTC") {
		ftime = ftime[0:len(ftime)-3] + "GMT"
	}
	return ftime
}

func GetLocalAddr() string {
	info, _ := net.InterfaceAddrs()
	for _, addr := range info {
		ip := strings.Split(addr.String(), "/")[0]
		if ip != "0.0.0.0" {
			return ip
		}
	}
	return ""
}

//不创建新数组切片，只在数组内部修改值
func BytesToLower(bs []byte) {
	for i := 0; i < len(bs); i++ {
		if 'A' <= bs[i] && bs[i] <= 'Z' {
			bs[i] += ('a' - 'A')
		}
	}
}

func BytesToUpper(bs []byte) {
	for i := 0; i < len(bs); i++ {
		if 'a' <= bs[i] && bs[i] <= 'z' {
			bs[i] -= ('a' - 'A')
		}
	}
}
