package config

import "testing"

func Test_readCofigFromYamlFile(t *testing.T) {
	type args struct {
		file string
		body any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				file: "test.yaml",
				body: &struct {
					Name string
				}{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ReadYamlFile(tt.args.file, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("readCofigFromYamlFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
