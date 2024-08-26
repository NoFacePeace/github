package linkedlist

import (
	"reflect"
	"testing"
)

func TestRandomizedSet(t *testing.T) {
	tests := []struct {
		name string
		want RandomizedSet
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRandomizedSet(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRandomizedSet() = %v, want %v", got, tt.want)
			}
		})
	}
}
