package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMemoryStoreWritesAndRebuildsIndex(t *testing.T) {
	store := newMemoryStore(t.TempDir())
	if err := store.ensure(); err != nil {
		t.Fatal(err)
	}
	if err := store.write(extractedMemory{Name: "User Preference/Tabs", Type: "user", Description: "Prefer tabs", Body: "Use tabs for indentation."}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(store.dir, "user-preference-tabs.md")); err != nil {
		t.Fatal(err)
	}
	if got := store.index(); !strings.Contains(got, "[User Preference/Tabs](user-preference-tabs.md) — Prefer tabs") {
		t.Fatalf("unexpected index: %q", got)
	}
	files, err := store.list()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Body != "Use tabs for indentation." {
		t.Fatalf("unexpected files: %#v", files)
	}
}

func TestSafePathRejectsEscapingWorkspace(t *testing.T) {
	if _, err := safePath("../../outside"); err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
}
