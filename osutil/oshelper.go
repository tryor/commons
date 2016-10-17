package osutil

import (
	"os"
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
