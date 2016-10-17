package ecutil

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"code.google.com/p/go.text/encoding/simplifiedchinese"
)

func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func GBKToUtf8(src string) (string, error) {
	res := make([]byte, len(src)*2)
	decder := simplifiedchinese.GBK.NewDecoder()
	nDst, _, err := decder.Transform(res, []byte(src), true)
	if err != nil {
		return "", err
	}
	return string(res[0:nDst]), nil
}

func Utf8ToGBK(src string) (string, error) {
	res := make([]byte, len(src)*2)
	decder := simplifiedchinese.GBK.NewEncoder()
	nDst, _, err := decder.Transform(res, []byte(src), true)
	if err != nil {
		return "", err
	}
	return string(res[0:nDst]), nil
}
