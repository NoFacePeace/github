package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRushFlowOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{
		"templateid": "template-test",
		"collecttype": 20,
		"mobile": "13800000000"
	}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	got, err := loadRushFlowOptions(path)
	if err != nil {
		t.Fatalf("loadRushFlowOptions() error = %v", err)
	}
	if got.TemplateID != "template-test" {
		t.Errorf("TemplateID = %q", got.TemplateID)
	}
	if got.CollectType != 20 {
		t.Errorf("CollectType = %d", got.CollectType)
	}
	if got.Mobile != "13800000000" {
		t.Errorf("Mobile = %q", got.Mobile)
	}
}

func TestLoadRushFlowOptionsDefaultCollectType(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	got, err := loadRushFlowOptions(path)
	if err != nil {
		t.Fatalf("loadRushFlowOptions() error = %v", err)
	}
	if got.CollectType != 10 {
		t.Errorf("CollectType = %d, want 10", got.CollectType)
	}
}
