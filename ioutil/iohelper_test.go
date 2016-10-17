package ioutil

import (
	"net/url"
	"testing"
)

func Test_GetHttpClient(t *testing.T) {
	resp, err := DefaultClient.Get("https://www.baidu.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.Status, resp.ContentLength)
	t.Log(resp.Body)
	t.Log(resp.Close)
}

func Test_HttpGet(t *testing.T) {
	code, status, body, err := HttpGet("http://www.oschina.net/")
	t.Log(code, status, len(body), err)
	//t.Log(body)
}

func Test_HttpPostForm(t *testing.T) {
	var data url.Values
	code, status, body, err := HttpPostForm("http://www.baidu.com", data)
	t.Log(code, status, len(body), err)
	//t.Log(body)
}
