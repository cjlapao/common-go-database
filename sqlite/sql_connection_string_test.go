package sqlite

import "testing"

func TestSqlConnectionString_Parse(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		c       *SqlConnectionString
		args    args
		wantErr bool
	}{
		{
			"valid path",
			&SqlConnectionString{
				FilePath: "./",
				FileName: "test.db",
			},
			args{
				"test",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Parse(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("SqlConnectionString.Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
