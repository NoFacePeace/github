package str

import "testing"

func Test_validateCoupons(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		code         []string
		businessLine []string
		isActive     []bool
		want         []string
	}{
		{
			businessLine: []string{"restaurant", "grocery", "pharmacy", "restaurant"},
			code:         []string{"SAVE20", "", "PHARMA5", "SAVE@20"},
			isActive:     []bool{true, true, true, true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateCoupons(tt.code, tt.businessLine, tt.isActive)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("validateCoupons() = %v, want %v", got, tt.want)
			}
		})
	}
}
