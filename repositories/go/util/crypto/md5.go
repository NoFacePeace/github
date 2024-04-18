package crypto

import (
	"crypto/md5"
	"encoding/json"
	"io"
)

func Md5Sum(data any) (string, error) {
	bs, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	h := md5.New()
	_, err = io.WriteString(h, string(bs))
	if err != nil {
		return "", err
	}
	return string(h.Sum(nil)), nil
}
