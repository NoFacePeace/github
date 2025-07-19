package seatalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Interface interface {
	SendGroupText(string, string) error
}

type SeaTalk struct {
	url string
}

func New(url string) Interface {
	return &SeaTalk{
		url: url,
	}
}

func (s *SeaTalk) SendGroupText(group, text string) error {
	req := &SendGroupTextReq{}
	req.Tag = "text"
	req.Text.Content = text
	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("json marshal error: [%w]", err)
	}
	endpoint := s.url + "/webhook/group/" + group
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("http post error: [%w]", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io read all error: [%w]", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status code error: [%s]", string(body))
	}
	ret := &SendGroupTextResp{}
	if err := json.Unmarshal(body, ret); err != nil {
		return fmt.Errorf("json unmarshal error: [%w]", err)
	}
	if ret.Code != 0 {
		return fmt.Errorf("seatalk send group text error: [%s]", ret.Msg)
	}
	return nil
}

type SendGroupTextReq struct {
	Tag  string `json:"tag"`
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
}

type SendGroupTextResp struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	MessageID string `json:"message_id"`
}
