package ismtp

import (
	"crypto/tls"
	"encoding/base64"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"
)

//基于GO1.7修改
//1. 增加 DialTimeout(), 可创建带超时机制连接
//2. 修改Auth(), 解决对某些smtp不兼容问题
//3. 修改SendMail(), 增加timeout参数

func DialTimeout(addr string, timeout time.Duration) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(timeout))
	host := addr[:strings.Index(addr, ":")]
	return NewClient(conn, host)
}

func (c *Client) Auth(a smtp.Auth) error {
	if err := c.hello(); err != nil {
		return err
	}
	encoding := base64.StdEncoding
	mech, resp, err := a.Start(&smtp.ServerInfo{c.serverName, c.tls, c.auth})
	if err != nil {
		c.Quit()
		return err
	}

	var code int
	var msg64 string
	var resp64 []byte
	if len(resp) > 0 {
		resp64 = make([]byte, encoding.EncodedLen(len(resp)))
		encoding.Encode(resp64, resp)
		code, msg64, err = c.cmd(0, "AUTH %s %s", mech, resp64)
	} else {
		code, msg64, err = c.cmd(0, "AUTH %s", mech)
	}

	for err == nil {
		var msg []byte
		switch code {
		case 334:
			msg, err = encoding.DecodeString(msg64)
		case 235:
			// the last message isn't base64 because it isn't a challenge
			msg = []byte(msg64)
		default:
			err = &textproto.Error{Code: code, Msg: msg64}
		}
		if err == nil {
			resp, err = a.Next(msg, code == 334)
		}
		if err != nil {
			// abort the AUTH
			c.cmd(501, "*")
			c.Quit()
			break
		}
		if resp == nil {
			break
		}
		resp64 = make([]byte, encoding.EncodedLen(len(resp)))
		encoding.Encode(resp64, resp)
		code, msg64, err = c.cmd(0, string(resp64))
	}
	return err
}

func SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte, timeout ...time.Duration) error {
	var c *Client
	var err error
	if len(timeout) > 0 {
		c, err = DialTimeout(addr, timeout[0])
	} else {
		c, err = Dial(addr)
	}
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.hello(); err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: c.serverName}
		if testHookStartTLS != nil {
			testHookStartTLS(config)
		}
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}
	if a != nil && c.ext != nil {
		if _, ok := c.ext["AUTH"]; ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
