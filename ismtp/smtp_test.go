package ismtp

import (
	"fmt"
	"time"
	//"net/smtp"
	"testing"
)

func test_ismtp(t *testing.T) {
	auth := LoginAuth(
		"trywen001@126.com",
		"123qwe",
		"smtp.126.com",
	)

	ctype := fmt.Sprintf("Content-Type: %s; charset=%s", "text/html", "utf-8")
	msg := fmt.Sprintf("To: %s\r\nCc: %s\r\nFrom: %s\r\nSubject: %s\r\n%s\r\n\r\n%s", "trywen@qq.com", "", "trywen001@126.com", "测试(Hello)", ctype, "<html><body>Hello Hello</body></html>")

	err := SendMail(
		"smtp.126.com:25",
		auth,
		"trywen001@126.com",
		[]string{"trywen@qq.com"},
		[]byte(msg),
		time.Second*10,
	)
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}
}

func Test_ismtp(t *testing.T) {
	auth := LoginAuth(
		"wenjian@tv189.com",
		"************",
		"smtp.tv189.com",
	)

	ctype := fmt.Sprintf("Content-Type: %s; charset=%s", "text/html", "utf-8")
	msg := fmt.Sprintf("To: %s\r\nCc: %s\r\nFrom: %s\r\nSubject: %s\r\n%s\r\n\r\n%s", "trywen@qq.com", "", "trywen001@126.com", "测试(Hello)", ctype, "<html><body>Hello Hello</body></html>")

	err := SendMail(
		"smtp.tv189.com:25",
		auth,
		"wenjian@tv189.com",
		[]string{"trywen@qq.com"},
		[]byte(msg),
		time.Second*10,
	)
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}
}
