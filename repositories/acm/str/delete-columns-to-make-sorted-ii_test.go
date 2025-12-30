package str

import "testing"

func Test_minDeletionSizeII(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		strs []string
		want int
	}{
		{
			strs: []string{"ffusbkyqlb", "ercqbqkrhb", "tjghblnrtn", "soflcftrsy", "afexdrmbxo", "zvotdsjiyg", "tosldognaf", "vgrugbnqre", "ohpchuqazm", "lsgjitblxb", "oemujbxnxm", "nywzjglrug", "ermokiwkdi", "cnzykvhyci", "fdsblgitww", "esofnnmnhs", "lawlnyuwwx", "gijnnhtydz", "lqfkqmlcnn", "mchvrcovml", "slatswujew", "krebwrebsj", "kapfwsvmvv", "tzuawyxsqu", "aiuqwtuzdw", "ynkrxfehjc", "nkuuyqsire", "fktpymcvmr", "xxkfygbzzv", "oiaxzreocg"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minDeletionSizeII(tt.strs)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("minDeletionSizeII() = %v, want %v", got, tt.want)
			}
		})
	}
}
