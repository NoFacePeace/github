package workspaces

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDirectory(t *testing.T) {
	cfg := Config{DataDirectory: filepath.Join(t.TempDir(), "data")}
	tests := []struct {
		name        string
		workspaceID string
		directory   string
	}{
		{name: "unset", directory: "default"},
		{name: "default", workspaceID: "default", directory: "default"},
		{name: "custom", workspaceID: "workspace_123", directory: "workspace_123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := filepath.Join(cfg.DataDirectory, "workspaces", tt.directory)
			for range 2 {
				directory, err := EnsureDirectory(cfg, tt.workspaceID)
				if err != nil {
					t.Fatalf("EnsureDirectory() error = %v", err)
				}
				if directory != want {
					t.Errorf("EnsureDirectory() directory = %q, want %q", directory, want)
				}
				info, err := os.Stat(want)
				if err != nil {
					t.Fatalf("Stat(%q) error = %v", want, err)
				}
				if !info.IsDir() {
					t.Errorf("EnsureDirectory() did not create a directory at %q", want)
				}
			}
		})
	}
}

func TestEnsureDirectoryRejectsInvalidPaths(t *testing.T) {
	dataDirectory := filepath.Join(t.TempDir(), "data")
	tests := []struct {
		name          string
		dataDirectory string
		workspaceID   string
	}{
		{name: "empty data directory"},
		{name: "relative data directory", dataDirectory: "data"},
		{name: "current directory id", dataDirectory: dataDirectory, workspaceID: "."},
		{name: "parent directory id", dataDirectory: dataDirectory, workspaceID: ".."},
		{name: "traversal id", dataDirectory: dataDirectory, workspaceID: "../outside"},
		{name: "unclean id", dataDirectory: dataDirectory, workspaceID: "nested/../other"},
		{name: "absolute id", dataDirectory: dataDirectory, workspaceID: filepath.Join(dataDirectory, "other")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directory, err := EnsureDirectory(Config{DataDirectory: tt.dataDirectory}, tt.workspaceID)
			if err == nil {
				t.Fatal("EnsureDirectory() error = nil, want error")
			}
			if directory != "" {
				t.Errorf("EnsureDirectory() directory = %q, want empty", directory)
			}
		})
	}
	if _, err := os.Stat(dataDirectory); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Stat(%q) error = %v, want os.ErrNotExist", dataDirectory, err)
	}
}

func TestEnsureDirectoryReturnsFilesystemError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "data")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	directory, err := EnsureDirectory(Config{DataDirectory: file}, "")
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("EnsureDirectory() error = %v, want wrapped *os.PathError", err)
	}
	if directory != "" {
		t.Errorf("EnsureDirectory() directory = %q, want empty", directory)
	}
}

func TestValidateDirectory(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "file.txt")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(directory, "missing")
	tests := []struct {
		name         string
		directory    string
		wantErr      bool
		wantNotExist bool
	}{
		{name: "directory", directory: directory},
		{name: "empty", wantErr: true},
		{name: "relative", directory: "workspace", wantErr: true},
		{name: "missing", directory: missing, wantErr: true, wantNotExist: true},
		{name: "file", directory: file, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectory(tt.directory)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateDirectory(%q) error = %v, wantErr %v", tt.directory, err, tt.wantErr)
			}
			if tt.wantNotExist && !errors.Is(err, os.ErrNotExist) {
				t.Errorf("ValidateDirectory(%q) error = %v, want wrapped os.ErrNotExist", tt.directory, err)
			}
		})
	}
}

func TestValidateDirectorySymlinks(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "file.txt")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{name: "directory", target: directory},
		{name: "file", target: file, wantErr: true},
		{name: "broken", target: filepath.Join(directory, "missing"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := filepath.Join(t.TempDir(), "link")
			if err := os.Symlink(tt.target, link); err != nil {
				t.Fatal(err)
			}
			err := ValidateDirectory(link)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateDirectory(%q) error = %v, wantErr %v", link, err, tt.wantErr)
			}
		})
	}
}
