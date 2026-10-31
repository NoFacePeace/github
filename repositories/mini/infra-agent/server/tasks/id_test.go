package tasks

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewID(t *testing.T) {
	for _, prefix := range []string{"task", "run", "attachment", "executor"} {
		t.Run(prefix, func(t *testing.T) {
			id, err := newID(prefix)
			if err != nil {
				t.Fatalf("newID() error = %v", err)
			}
			assertPrefixedUUIDv7(t, id, prefix)
		})
	}
}

func assertPrefixedUUIDv7(t *testing.T, id, prefix string) {
	t.Helper()
	value, ok := strings.CutPrefix(id, prefix+"_")
	if !ok {
		t.Fatalf("ID = %q, want prefix %q", id, prefix+"_")
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("UUID = %q, parse error = %v", value, err)
	}
	if parsed.Version() != 7 {
		t.Errorf("UUID version = %v, want 7", parsed.Version())
	}
	if parsed.String() != value {
		t.Errorf("UUID = %q, want canonical format %q", value, parsed.String())
	}
}
