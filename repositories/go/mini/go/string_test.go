package mini

import "testing"

func TestConcatStrings(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{name: "empty", want: ""},
		{name: "skip empty", parts: []string{"", "Go", "", " string"}, want: "Go string"},
		{name: "single", parts: []string{"hello"}, want: "hello"},
		{name: "large", parts: []string{"abcdefghijklmnopqrstuvwxyz", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "0123456789"}, want: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConcatStrings(tt.parts...); got != tt.want {
				t.Fatalf("ConcatStrings() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBytesToString(t *testing.T) {
	source := []byte("hello")
	got := BytesToString(source)
	source[0] = 'j'

	if got != "hello" {
		t.Fatalf("BytesToString() result changed to %q after source mutation", got)
	}
}

func TestStringToBytes(t *testing.T) {
	source := "hello"
	got := StringToBytes(source)
	got[0] = 'j'

	if source != "hello" {
		t.Fatalf("StringToBytes() changed source to %q", source)
	}
	if string(got) != "jello" {
		t.Fatalf("StringToBytes() = %q, want %q", got, "jello")
	}
}
