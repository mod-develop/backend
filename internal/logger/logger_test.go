package logger

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		options []option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "new",
			args: args{
				options: []option{
					SetEnableFileOutput(false),
					SetEnableTerminalOutput(false),
					SetLevel("info"),
					SetLogPath("./testdata/log.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "bad level",
			args: args{
				options: []option{
					SetEnableFileOutput(false),
					SetEnableTerminalOutput(false),
					SetLevel("info123"),
					SetLogPath("./testdata/log.log"),
				},
			},
			wantErr: true,
		},
		{
			name: "bad logpath",
			args: args{
				options: []option{
					SetEnableFileOutput(false),
					SetEnableTerminalOutput(false),
					SetLevel("info123"),
					SetLogPath(""),
				},
			},
			wantErr: true,
		},
	}
	_ = os.RemoveAll("./testdata")
	_ = os.Remove("./logs")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
