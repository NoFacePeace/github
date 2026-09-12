package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type rawPacket struct {
	URL      string `json:"url"`
	Method   string `json:"method"`
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Path     string `json:"path"`
	Req      struct {
		Method  string                     `json:"method"`
		Headers map[string]json.RawMessage `json:"headers"`
		Body    string                     `json:"body"`
		Base64  string                     `json:"base64"`
	} `json:"req"`
}

func loadPacketRequest(path string) (requestParams, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return requestParams{}, err
	}

	var packet rawPacket
	if err := json.Unmarshal(data, &packet); err != nil {
		return requestParams{}, err
	}

	headers := make(http.Header, len(packet.Req.Headers))
	host := ""
	for name, value := range packet.Req.Headers {
		values, err := decodeHeaderValues(value)
		if err != nil {
			return requestParams{}, fmt.Errorf("decode header %q: %w", name, err)
		}

		if strings.EqualFold(name, "Host") {
			if len(values) > 0 {
				host = values[0]
			}
			continue
		}
		if strings.EqualFold(name, "Content-Length") {
			continue
		}

		for _, value := range values {
			headers.Add(name, value)
		}
	}

	body, err := decodePacketBody(packet.Req.Body, packet.Req.Base64)
	if err != nil {
		return requestParams{}, err
	}

	method := packet.Req.Method
	if method == "" {
		method = packet.Method
	}
	if method == "" {
		method = http.MethodPost
	}

	requestURL := packet.URL
	if requestURL == "" {
		scheme := strings.ToLower(packet.Protocol)
		if scheme == "" {
			scheme = "https"
		}
		requestURL = scheme + "://" + packet.Hostname + packet.Path
	}

	return requestParams{
		URL:     requestURL,
		Method:  method,
		Host:    host,
		Headers: headers,
		Body:    body,
	}, nil
}

func decodeHeaderValues(raw json.RawMessage) ([]string, error) {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return []string{single}, nil
	}

	var multiple []string
	if err := json.Unmarshal(raw, &multiple); err == nil {
		return multiple, nil
	}

	return nil, errors.New("expected a string or string array")
}

func decodePacketBody(body, encodedBody string) ([]byte, error) {
	if encodedBody == "" {
		return []byte(body), nil
	}

	decoded, err := base64.StdEncoding.DecodeString(encodedBody)
	if err != nil {
		return nil, fmt.Errorf("decode request body: %w", err)
	}
	return decoded, nil
}
