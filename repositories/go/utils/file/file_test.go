package file

import "testing"

func TestPersistMetaData(t *testing.T) {
	type args struct {
		fileName string
		data     []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				fileName: "test",
				data:     []byte("test"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PersistMetaData(tt.args.fileName, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("PersistMetaData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
