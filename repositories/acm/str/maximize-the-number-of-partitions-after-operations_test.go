package str

import "testing"

func Test_maxPartitionsAfterOperations(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    string
		k    int
		want int
	}{
		{
			s: "noyynxgvtkhxsqdqcjyecjpwcawkgsrxmixokubliztvglyftkcrkpdfofwhaydetelrlyzirwmcjlnghqzsepsztnshfsanwezyrwugjtupaukeqhnqjuuyzlixhzewymafxyjasqlfvvabungssaylgcxydwvnwcayoogevdkpkxbvofwgohtjocqhtykbrpurqxqvwyxdxxqhstlbkcuohtkmlyqfdzcbatmshcpoeoqirqtyuabiwrtyprucmfpcezmawmjhsskexpzlnasejilkjtbwuylzdpunifykhyteoglauzfaljvndlpeubkxtmnisawrdlzfcvfljdrtnzwhyuelqdtbgjvrublexxslrckupnwznerwanngvfppxnayeorsgnozapmgnsbzuxmaeoyrfwhhsdnxsflqklbtopradhxgadzjrrdutduhiurdjaovkgtulcjndpcibywdzwxucxakouievplehkdkdhpnfgjqrrjcwdnwgfujzpkihjjvxrdtluuxdpzwwgdifhzvuuhpoe",
			k: 22,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxPartitionsAfterOperations(tt.s, tt.k)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("maxPartitionsAfterOperations() = %v, want %v", got, tt.want)
			}
		})
	}
}
