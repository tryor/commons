package ismtp

import (
	"fmt"
	"testing"
)

func Test_ismtp(t *testing.T) {
	auth := LoginAuth(
		"trywen001@126.com",
		"123qwe",
		"smtp.126.com",
	)

	ctype := fmt.Sprintf("Content-Type: %s; charset=%s", "text/html", "utf-8")
	msg := fmt.Sprintf("To: %s\r\nCc: %s\r\nFrom: %s\r\nSubject: %s\r\n%s\r\n\r\n%s", "<TTT>trywen@qq.com", "", "TRY<trywen001@126.com>", "Hello", ctype, "<html><body>Hello Hello</body></html>")

	err := SendMail(
		"smtp.126.com:25",
		auth,
		"trywen001@126.com",
		[]string{"trywen@qq.com"},
		[]byte(msg),
	)
	if err != nil {
		t.Error(err)
	}
}
