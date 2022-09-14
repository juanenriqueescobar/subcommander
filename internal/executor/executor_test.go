package executor

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

// verifica q no se reviente cuando una linea de log pesa 1MB
func Test_exec_very_long_log(t *testing.T) {
	l := logrus.New()
	l.SetLevel(logrus.ErrorLevel)
	type args struct {
		id   string
		data StdinData
	}
	tests := []struct {
		name string
		ex   *Executor
		args args
		want bool
	}{
		{
			name: "1",
			ex:   NewExec("php", []string{"testing/tester_2.php"}, time.Hour, l.WithField("xx", "yy")),
			args: args{
				id:   "1",
				data: StdinData{Body: `{"order_id":1}`, Metadata: map[string]interface{}{"trace_id": "aabbcc"}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.ex.Run(tt.args.id, tt.args.data); got != tt.want {
				t.Errorf("exec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_exec_ok(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	type args struct {
		id   string
		data StdinData
	}
	tests := []struct {
		name string
		ex   *Executor
		args args
		want bool
	}{
		{
			name: "1",
			ex:   NewExec("php", []string{"testing/tester_1.php"}, time.Hour, logrus.NewEntry(logger)),
			args: args{
				id:   "1",
				data: StdinData{Body: `{"order_id":1}`, Metadata: map[string]interface{}{"trace_id": "aabbcc"}},
			},
			want: true,
		},
		{
			name: "2",
			ex:   NewExec("php", []string{"testing/tester_1.php", "--xxx"}, time.Hour, logrus.NewEntry(logger)),
			args: args{
				id:   "2",
				data: StdinData{Body: `{"order_id":1,"abc":"cba"}`, Metadata: map[string]interface{}{"trace_id": "aabbcc", "open_id": "112233"}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.ex.Run(tt.args.id, tt.args.data); got != tt.want {
				t.Errorf("exec() = %v, want %v", got, tt.want)
			}

			body := map[string]interface{}{}
			err := json.Unmarshal([]byte(tt.args.data.Body), &body)
			if err != nil {
				t.Error(err)
			}

			l := hook.AllEntries()

			assert.Equal(t, l[0].Data, logrus.Fields{
				"event.category":       []string{"process"},
				"event.dataset":        "subcommander.exec",
				"event.type":           []string{"start"},
				"process.args":         tt.ex.args,
				"process.args_count":   len(tt.ex.args),
				"process.command_line": tt.ex.command,
				"transaction.id":       tt.args.id,

				"sc_task.args":    tt.ex.args,
				"sc_task.cmd":     tt.ex.command,
				"sc_task.action":  "started",
				"sc_task.started": 1,
				"sc_task.id":      tt.args.id,
			})
			assert.Equal(t, l[0].Message, "task start")
			assert.Equal(t, l[0].Level, logrus.InfoLevel)

			//-----
			assert.Equal(t, l[1].Data, logrus.Fields{
				"event.category":       []string{"process"},
				"event.dataset":        "subcommander.exec",
				"process.args":         tt.ex.args,
				"process.args_count":   len(tt.ex.args),
				"process.command_line": tt.ex.command,
				"transaction.id":       tt.args.id,

				"sc_task.id": tt.args.id,
			})
			assert.Equal(t, l[1].Message, "step1")
			assert.Equal(t, l[1].Level, logrus.DebugLevel)

			assert.Equal(t, l[2].Data, logrus.Fields{
				"event.category":       []string{"process"},
				"event.dataset":        "subcommander.exec",
				"process.args":         tt.ex.args,
				"process.args_count":   len(tt.ex.args),
				"process.command_line": tt.ex.command,
				"transaction.id":       tt.args.id,

				"sc_task.id": tt.args.id,
			})
			assert.Equal(t, l[2].Message, "step2")
			assert.Equal(t, l[2].Level, logrus.DebugLevel)

			assert.Equal(t, l[3].Data, logrus.Fields{
				"event.category":       []string{"process"},
				"event.dataset":        "subcommander.exec",
				"process.args":         tt.ex.args,
				"process.args_count":   len(tt.ex.args),
				"process.command_line": tt.ex.command,
				"transaction.id":       tt.args.id,

				"metadata":   tt.args.data.Metadata,
				"sc_task.id": tt.args.id,
			})
			assert.Equal(t, l[3].Message, "step3")
			assert.Equal(t, l[3].Level, logrus.InfoLevel)

			assert.Equal(t, l[4].Data, logrus.Fields{
				"event.category":       []string{"process"},
				"event.dataset":        "subcommander.exec",
				"process.args":         tt.ex.args,
				"process.args_count":   len(tt.ex.args),
				"process.command_line": tt.ex.command,
				"transaction.id":       tt.args.id,

				"body":       body,
				"sc_task.id": tt.args.id,
			})
			assert.Equal(t, l[4].Message, "step4")
			assert.Equal(t, l[4].Level, logrus.WarnLevel)

			assert.Equal(t, l[5].Data, logrus.Fields{
				"event.category":       []string{"process"},
				"event.dataset":        "subcommander.exec",
				"process.args":         tt.ex.args,
				"process.args_count":   len(tt.ex.args),
				"process.command_line": tt.ex.command,
				"transaction.id":       tt.args.id,

				"sc_task.id": tt.args.id,
			})
			assert.Equal(t, l[5].Message, "step5")
			assert.Equal(t, l[5].Level, logrus.ErrorLevel)

			assert.Equal(t, l[6].Data, logrus.Fields{
				"event.category":       []string{"process"},
				"event.dataset":        "subcommander.exec",
				"process.args":         tt.ex.args,
				"process.args_count":   len(tt.ex.args),
				"process.command_line": tt.ex.command,
				"transaction.id":       tt.args.id,

				"sc_task.id": tt.args.id,
			})
			assert.Equal(t, l[6].Message, "step6")
			assert.Equal(t, l[6].Level, logrus.DebugLevel)

			assert.Equal(t, l[7].Data, logrus.Fields{
				"event.category":       []string{"process"},
				"event.dataset":        "subcommander.exec",
				"event.duration":       l[7].Data["event.duration"],
				"event.outcome":        "success",
				"event.type":           []string{"end"},
				"process.args":         tt.ex.args,
				"process.args_count":   len(tt.ex.args),
				"process.command_line": tt.ex.command,
				"process.exit_code":    0,
				"transaction.id":       tt.args.id,

				"sc_task.duration": l[7].Data["sc_task.duration"],
				"sc_task.action":   "finished",
				"sc_task.finished": 1,
				"sc_task.status":   "success",
				"sc_task.success":  1,
				"sc_task.id":       tt.args.id,
			})
			assert.Equal(t, l[7].Message, "task finish")
			assert.Equal(t, l[7].Level, logrus.InfoLevel)

			hook.Reset()
		})
	}
}

func Test_exec_fail(t *testing.T) {

	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	type args struct {
		id   string
		data StdinData
	}
	tests := []struct {
		name string
		ex   *Executor
		args args
		want bool
	}{
		{
			name: "1",
			ex:   NewExec("bash", []string{"testing/sleep.sh"}, 1*time.Second, logrus.NewEntry(logger)),
			args: args{
				id:   "1",
				data: StdinData{Body: `{"order_id":1}`, Metadata: map[string]interface{}{"trace_id": "aabbcc"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.ex.Run(tt.args.id, tt.args.data); got != tt.want {
				t.Errorf("exec() = %v, want %v", got, tt.want)
			}

			body := map[string]interface{}{}
			err := json.Unmarshal([]byte(tt.args.data.Body), &body)
			if err != nil {
				t.Error(err)
			}

			l := hook.AllEntries()

			assert.Equal(t, l[0].Data, logrus.Fields{
				"event.category":       []string{"process"},
				"event.dataset":        "subcommander.exec",
				"event.type":           []string{"start"},
				"process.args":         tt.ex.args,
				"process.args_count":   len(tt.ex.args),
				"process.command_line": tt.ex.command,
				"transaction.id":       tt.args.id,

				"sc_task.args":    tt.ex.args,
				"sc_task.cmd":     tt.ex.command,
				"sc_task.action":  "started",
				"sc_task.started": 1,
				"sc_task.id":      tt.args.id,
			})
			assert.Equal(t, l[0].Message, "task start")
			assert.Equal(t, l[0].Level, logrus.InfoLevel)

			assert.Equal(t, l[1].Data, logrus.Fields{
				"event.category":       []string{"process"},
				"event.dataset":        "subcommander.exec",
				"event.duration":       l[1].Data["event.duration"],
				"event.outcome":        "failure",
				"event.type":           []string{"end"},
				"process.args":         tt.ex.args,
				"process.args_count":   len(tt.ex.args),
				"process.command_line": tt.ex.command,
				"process.exit_code":    -1,
				"transaction.id":       tt.args.id,

				"error": errors.New("signal: terminated"),

				"sc_task.duration": l[1].Data["sc_task.duration"],
				"sc_task.action":   "finished",
				"sc_task.finished": 1,
				"sc_task.status":   "fail",
				"sc_task.fail":     1,
				"sc_task.id":       tt.args.id,
			})
			assert.Equal(t, l[1].Message, "task finish")
			assert.Equal(t, l[1].Level, logrus.ErrorLevel)

			hook.Reset()
		})
	}
}
