package datetime

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	path = "https://raw.githubusercontent.com/NoFacePeace/github/master/repositories/go/utils/datetime/"
)

func IsHoliday(t time.Time) (bool, error) {
	year := t.Year()
	url := fmt.Sprintf(path+"%v.txt", year)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	arr := strings.Split(string(body), "\n")
	for _, v := range arr {
		if v == "" {
			continue
		}
		h, err := time.Parse(LayoutDate, v)
		if err != nil {
			return false, err
		}
		if EqualDate(t, h) {
			return true, nil
		}
	}
	return false, nil
}
