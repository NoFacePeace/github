package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJobJSON(t *testing.T) {
	want := Job{MessageID: "job-1", RunnerID: "runner-1", LeaseID: "lease-1", LeaseExpiresAt: time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC)}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["message_id"]) != `"job-1"` || fields["MessageID"] != nil {
		t.Fatalf("unexpected message ID field in %s", data)
	}
	if string(fields["lease_id"]) != `"lease-1"` || string(fields["lease_expires_at"]) != `"2026-11-28T00:00:00Z"` {
		t.Fatalf("unexpected lease fields in %s", data)
	}
	var roundTrip Job
	if err := json.Unmarshal(data, &roundTrip); err != nil || roundTrip != want {
		t.Fatalf("decode new job = %+v, %v; want %+v", roundTrip, err, want)
	}
}
