package ioutil

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

const READ_BUFFER_SIZE = 1024 * 10

const HTTP_TIME_OUT = 10 //默认IO超时时间，单位：秒
var DefaultClient *http.Client

func init() {
	DefaultClient = GetHttpClient(HTTP_TIME_OUT)
}

//timeout 单位秒
func GetHttpClient(timeout time.Duration) *http.Client {
	c := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(timeout * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*timeout)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
			ResponseHeaderTimeout: time.Second * 5,
		},
	}
	return c
}

func HttpPost(sendurl string, reqbody io.Reader, bodytype string, client ...*http.Client) (code int, status string, body string, err error) {
	var resp *http.Response
	var httpClient *http.Client
	if len(client) > 0 {
		httpClient = client[0]
	} else {
		httpClient = DefaultClient
	}
	resp, err = httpClient.Post(sendurl, bodytype, reqbody)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		return
	}

	code = resp.StatusCode
	status = resp.Status

	if resp.Body != nil {
		var data []byte
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		body = string(data)
	}

	return
}

func HttpPostForm(sendurl string, data url.Values, client ...*http.Client) (code int, status string, body string, err error) {
	var resp *http.Response
	var httpClient *http.Client
	if len(client) > 0 {
		httpClient = client[0]
	} else {
		httpClient = DefaultClient
	}
	resp, err = httpClient.PostForm(sendurl, data)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		return
	}

	code = resp.StatusCode
	status = resp.Status

	if resp.Body != nil {
		var data []byte
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		body = string(data)
	}

	return
}

func HttpGet(url string, client ...*http.Client) (code int, status string, body string, err error) {
	var resp *http.Response
	var httpClient *http.Client
	if len(client) > 0 {
		httpClient = client[0]
	} else {
		httpClient = DefaultClient
	}
	resp, err = httpClient.Get(url)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		return
	}

	code = resp.StatusCode
	status = resp.Status
	//	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
	//		err = errors.New(strconv.Itoa(resp.StatusCode) + " " + resp.Status)
	//		return
	//	}
	if resp.Body != nil {
		var data []byte
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		body = string(data)

		//		var data *bytes.Buffer // []byte
		//		data, err = ReadAll(resp.Body)
		//		if err != nil {
		//			return
		//		}
		//		body = string(data.Bytes())
	}

	return
}

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
