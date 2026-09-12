package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPacketRequest(t *testing.T) {
	const requestBody = "encrypted-request-body"

	packetJSON := fmt.Sprintf(`{
		"url": "https://example.test/api",
		"method": "POST",
		"req": {
			"method": "POST",
			"headers": {
				"content-type": "application/json",
				"content-length": "22",
				"cookie": "session=test",
				"host": "example.test"
			},
			"base64": %q
		}
	}`, base64.StdEncoding.EncodeToString([]byte(requestBody)))

	packetPath := filepath.Join(t.TempDir(), "packet.json")
	if err := os.WriteFile(packetPath, []byte(packetJSON), 0o600); err != nil {
		t.Fatalf("write packet: %v", err)
	}

	got, err := loadPacketRequest(packetPath)
	if err != nil {
		t.Fatalf("loadPacketRequest() error = %v", err)
	}
	if got.URL != "https://example.test/api" {
		t.Errorf("URL = %q", got.URL)
	}
	if got.Host != "example.test" {
		t.Errorf("Host = %q", got.Host)
	}
	if got.Headers.Get("Cookie") != "session=test" {
		t.Errorf("Cookie = %q", got.Headers.Get("Cookie"))
	}
	if got.Headers.Get("Content-Length") != "" {
		t.Errorf("Content-Length should be derived from the body")
	}
	if string(got.Body) != requestBody {
		t.Errorf("Body = %q", got.Body)
	}
}
