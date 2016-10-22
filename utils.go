package util

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
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

//func Insert(slice []interface{}, index int, insertion ...interface{}) []interface{} {
//	result := make([]interface{}, len(slice)+len(insertion))
//	at := copy(result, slice[:index])
//	at += copy(result[at:], insertion)
//	copy(result[at:], slice[index:])
//	return result
//}

//body_buf := new(bytes.Buffer)

func ReadFull(r io.Reader) (*bytes.Buffer, error) {

	if _, ok := r.(*bufio.Reader); !ok {
		r = bufio.NewReader(r)
	}

	buffer := new(bytes.Buffer)

	bufsize := 1024
	buf := make([]byte, bufsize)

	for {
		//		n, err := io.ReadFull(r, buf)
		n, err := r.Read(buf)
		if n > 0 {
			buffer.Write(buf[0:n])
		}
		//		fmt.Println("n:", n)
		if err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				return buffer, err
			}
			break
		}
	}
	return buffer, nil
}
