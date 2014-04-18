package util

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func Test_BytesToLower(t *testing.T) {
	bs := []byte("ABcdEfg")
	BytesToLower(bs)
	//fmt.Println(string(bs))
	if string(bs) != "abcdefg" {
		t.Error("BytesToLower error")
	}

	bs = []byte("A中文Bcd汉E字fg")
	BytesToLower(bs)
	//fmt.Println(string(bs))
	if string(bs) != "a中文bcd汉e字fg" {
		t.Error("BytesToLower error")
	}

}

func Test_BytesToUpper(t *testing.T) {
	bs := []byte("ABcdEfg")
	BytesToUpper(bs)
	//fmt.Println(string(bs))
	if string(bs) != "ABCDEFG" {
		t.Error("BytesToUpper error")
	}

	bs = []byte("A中文Bcd汉E字fg")
	BytesToUpper(bs)
	//fmt.Println(string(bs))
	if string(bs) != "A中文BCD汉E字FG" {
		t.Error("BytesToUpper error")
	}
}

func Test_BytesToLower_UserTime(t *testing.T) {
	n := 10000000
	bs := make([]byte, n)
	for i := 0; i < n; i++ {
		bs[i] = 'A'
	}
	s := time.Now().UnixNano()
	BytesToLower(bs)
	fmt.Println(time.Now().UnixNano() - s)
}

func Test_bytes_ToLower_UserTime(t *testing.T) {
	n := 10000000
	bs := make([]byte, n)
	for i := 0; i < n; i++ {
		bs[i] = 'A'
	}
	s := time.Now().UnixNano()
	bytes.ToLower(bs)
	fmt.Println(time.Now().UnixNano() - s)
}
