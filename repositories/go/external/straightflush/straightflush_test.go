package straightflush

import (
	"context"
	"testing"
)

func Test_getBlockList(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		want    *getBlockListResp
		wantErr bool
	}{
		{
			name: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getBlockList(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("getBlockList() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("getBlockList() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("getBlockList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBlockList(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		want    []Block
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := GetBlockList(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetBlockList() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetBlockList() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetBlockList() = %v, want %v", got, tt.want)
			}
		})
	}
}
