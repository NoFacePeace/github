package slidingwindow

import "testing"

func Test_validSubstringCountII(t *testing.T) {
	type args struct {
		word1 string
		word2 string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			args: args{
				word1: "dddddededddeeeddd",
				word2: "eee",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validSubstringCountII(tt.args.word1, tt.args.word2); got != tt.want {
				t.Errorf("validSubstringCountII() = %v, want %v", got, tt.want)
			}
		})
	}
}
