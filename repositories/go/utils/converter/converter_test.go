package converter

import "testing"

func TestConvert(t *testing.T) {
	type args struct {
		a any
		b any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				a: &struct {
					Name string `json:"name"`
				}{Name: "test"},
				b: &map[string]string{
					"name": "test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Convert(tt.args.a, tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
