package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssembleSystemPromptSelectsMemoryOnlyWhenPresent(t *testing.T) {
	base := assembleSystemPrompt(promptContext{EnabledTools: []string{"bash"}, Workspace: "/work"})
	if strings.Contains(base, "Relevant memories") {
		t.Fatal("memory section should be absent")
	}
	withMemory := assembleSystemPrompt(promptContext{EnabledTools: []string{"bash"}, Workspace: "/work", Memories: "- [style](style.md) — tabs"})
	for _, want := range []string{"You are a coding agent", "Available tools: bash.", "Working directory: /work", "Relevant memories:"} {
		if !strings.Contains(withMemory, want) {
			t.Fatalf("prompt missing %q: %s", want, withMemory)
		}
	}
}

func TestUpdateContextReadsNonEmptyMemoryIndex(t *testing.T) {
	workdir := t.TempDir()
	if err := os.Mkdir(filepath.Join(workdir, ".memory"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workdir, ".memory", "MEMORY.md"), []byte("- [Go](go.md) — prefer standard library\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := updateContext(workdir); got.Memories == "" || got.Workspace != workdir {
		t.Fatalf("unexpected context: %#v", got)
	}
}

func TestSafePathRejectsTraversal(t *testing.T) {
	if _, err := safePath(t.TempDir(), "../../outside"); err == nil {
		t.Fatal("expected traversal rejection")
	}
}
