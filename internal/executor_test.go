package internal

import (
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func toMap(f logrus.Fields, n string) logrus.Fields {
	return f[n].(logrus.Fields)
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
			ex:   NewExec("php", []string{"testing/tester_1.php"}, logrus.NewEntry(logger)),
			args: args{
				id:   "1",
				data: StdinData{Body: `{"order_id":1}`, Metadata: map[string]interface{}{"trace_id": "aabbcc"}},
			},
			want: true,
		},
		{
			name: "2",
			ex:   NewExec("php", []string{"testing/tester_1.php", "--xxx"}, logrus.NewEntry(logger)),
			args: args{
				id:   "2",
				data: StdinData{Body: `{"order_id":1,"abc":"cba"}`, Metadata: map[string]interface{}{"trace_id": "aabbcc", "open_id": "112233"}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.ex.run(tt.args.id, tt.args.data); got != tt.want {
				t.Errorf("exec() = %v, want %v", got, tt.want)
			}

			body := map[string]interface{}{}
			err := json.Unmarshal([]byte(tt.args.data.Body), &body)
			if err != nil {
				t.Error(err)
			}

			l := hook.AllEntries()

			assert.Equal(t, l[0].Data, logrus.Fields{
				"sc_task": logrus.Fields{
					"args":    tt.ex.args,
					"cmd":     tt.ex.command,
					"action":  "started",
					"started": 1,
				},
				"sc_task.id": tt.args.id,
				"trace.id":   tt.args.id,
			})
			assert.Equal(t, l[0].Message, "task start")
			assert.Equal(t, l[0].Level, logrus.InfoLevel)

			assert.Equal(t, l[1].Data, logrus.Fields{
				"_jsonlog":   true,
				"sc_task.id": tt.args.id,
				"trace.id":   tt.args.id,
			})
			assert.Equal(t, l[1].Message, "step1")
			assert.Equal(t, l[1].Level, logrus.DebugLevel)

			assert.Equal(t, l[2].Data, logrus.Fields{
				"_jsonlog":   true,
				"sc_task.id": tt.args.id,
				"trace.id":   tt.args.id,
			})
			assert.Equal(t, l[2].Message, "step2")
			assert.Equal(t, l[2].Level, logrus.DebugLevel)

			assert.Equal(t, l[3].Data, logrus.Fields{
				"_jsonlog":   true,
				"metadata":   tt.args.data.Metadata,
				"sc_task.id": tt.args.id,
				"trace.id":   tt.args.id,
			})
			assert.Equal(t, l[3].Message, "step3")
			assert.Equal(t, l[3].Level, logrus.InfoLevel)

			assert.Equal(t, l[4].Data, logrus.Fields{
				"_jsonlog":   true,
				"body":       body,
				"sc_task.id": tt.args.id,
				"trace.id":   tt.args.id,
			})
			assert.Equal(t, l[4].Message, "step4")
			assert.Equal(t, l[4].Level, logrus.WarnLevel)

			assert.Equal(t, l[5].Data, logrus.Fields{
				"_jsonlog":   true,
				"sc_task.id": tt.args.id,
				"trace.id":   tt.args.id,
			})
			assert.Equal(t, l[5].Message, "step5")
			assert.Equal(t, l[5].Level, logrus.ErrorLevel)

			assert.Equal(t, l[6].Data, logrus.Fields{
				"_jsonlog":   false,
				"sc_task.id": tt.args.id,
				"trace.id":   tt.args.id,
			})
			assert.Equal(t, l[6].Message, "step6")
			assert.Equal(t, l[6].Level, logrus.DebugLevel)

			assert.Equal(t, l[7].Data, logrus.Fields{
				"sc_task": logrus.Fields{
					"duration": toMap(l[7].Data, "sc_task")["duration"].(int64), // Non deterministic
					"action":   "finished",
					"finished": 1,
				},
				"sc_task.success": 1,
				"sc_task.id":      tt.args.id,
				"trace.id":        tt.args.id,
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
			ex:   NewExec("php", []string{"testing/tester_0.php"}, logrus.NewEntry(logger)),
			args: args{
				id:   "1",
				data: StdinData{Body: `{"order_id":1}`, Metadata: map[string]interface{}{"trace_id": "aabbcc"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.ex.run(tt.args.id, tt.args.data); got != tt.want {
				t.Errorf("exec() = %v, want %v", got, tt.want)
			}

			body := map[string]interface{}{}
			err := json.Unmarshal([]byte(tt.args.data.Body), &body)
			if err != nil {
				t.Error(err)
			}

			l := hook.AllEntries()

			assert.Equal(t, l[0].Data, logrus.Fields{
				"sc_task": logrus.Fields{
					"args":    tt.ex.args,
					"cmd":     tt.ex.command,
					"action":  "started",
					"started": 1,
				},
				"sc_task.id": tt.args.id,
				"trace.id":   tt.args.id,
			})
			assert.Equal(t, l[0].Message, "task start")
			assert.Equal(t, l[0].Level, logrus.InfoLevel)

			assert.Equal(t, l[1].Data, logrus.Fields{
				"_jsonlog":   false,
				"sc_task.id": tt.args.id,
				"trace.id":   tt.args.id,
			})
			assert.Equal(t, l[1].Message, "Could not open input file: testing/tester_0.php")
			assert.Equal(t, l[1].Level, logrus.DebugLevel)

			assert.Equal(t, l[2].Data, logrus.Fields{
				"sc_task": logrus.Fields{
					"duration": toMap(l[2].Data, "sc_task")["duration"].(int64), // Non deterministic
					"action":   "finished",
					"finished": 1,
				},
				"sc_task.fail": 1,
				"sc_task.id":   tt.args.id,
				"trace.id":     tt.args.id,
			})
			assert.Equal(t, l[2].Message, "task finish")
			assert.Equal(t, l[2].Level, logrus.ErrorLevel)

			hook.Reset()
		})
	}
}
