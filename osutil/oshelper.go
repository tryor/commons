package osutil

import (
	"net"
	"os"
	"strings"
)

func MkdirAll(path string, mode ...os.FileMode) error {
	dir, err := os.Open(path)
	if err != nil {
		if len(mode) > 0 {
			return os.MkdirAll(path, mode[0])
		} else {
			return os.MkdirAll(path, os.ModePerm)
		}
	}
	if dir != nil {
		dir.Close()
	}
	return nil
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
