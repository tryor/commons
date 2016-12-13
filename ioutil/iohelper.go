package ioutil

import (
	"bufio"
	"bytes"
	"io"
)

const READ_BUFFER_SIZE = 1024 * 10

func ReadAll(r io.Reader) (*bytes.Buffer, error) {
	if _, ok := r.(*bufio.Reader); !ok {
		r = bufio.NewReader(r)
	}
	data := new(bytes.Buffer)
	//var data []byte
	buf := make([]byte, READ_BUFFER_SIZE)
	for {
		//n, err := io.ReadFull(r, buf)
		n, err := r.Read(buf)
		if n > 0 {
			data.Write(buf[0:n])
			//data = append(data, buf[0:n]...)
		}
		if err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				return data, err
			}
			break
		}
	}

	return data, nil
}
