package osutil

import (
	"log"
	//"os"
	"testing"
)

func Test_MkdirAll(t *testing.T) {
	err := MkdirAll("d:\\Temp\\aaaa\\bbbb\\file.txt")
	log.Println(err)
}
