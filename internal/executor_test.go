package internal

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func Test_exec(t *testing.T) {
	type args struct {
		id   string
		data string
	}
	tests := []struct {
		name string
		ex   *Executor
		args args
		want bool
	}{
		{
			name: "short",
			ex:   NewExec("php", []string{"testing/worker_1.php"}, logrus.New().WithField("", "")),
			args: args{
				id:   "1",
				data: `{"command":1}`,
			},
			want: true,
		},
		{
			name: "long process",
			ex:   NewExec("php", []string{"testing/worker_2.php"}, logrus.New().WithField("", "")),
			args: args{
				id:   "2",
				data: `{"order_id":1111}`,
			},
			want: true,
		},
		{
			name: "command not exists",
			ex:   NewExec("php", []string{".testing/worker_1.php"}, logrus.New().WithField("", "")),
			args: args{
				id:   "3",
				data: `{"command":1}`,
			},
			want: false,
		},
		{
			name: "command failed",
			ex:   NewExec("php", []string{"testing/worker_3.php"}, logrus.New().WithField("", "")),
			args: args{
				id:   "4",
				data: `{"command":1}`,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ex.run(tt.args.id, tt.args.data); got != tt.want {
				t.Errorf("exec() = %v, want %v", got, tt.want)
			}
		})
	}
	// t.Error(1)
}
