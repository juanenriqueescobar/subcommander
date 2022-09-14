package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/sirupsen/logrus"
)

type StdinData struct {
	Metadata map[string]interface{} `json:"metadata"`
	Body     string                 `json:"body"`
}

type Executor struct {
	parentLogger *logrus.Entry
	command      string
	args         []string
	timeout      time.Duration
}

func (e *Executor) logLine(log string, logger *logrus.Entry) {

	log = strings.TrimSpace(log)
	if len(log) == 0 {
		return
	}

	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(log), &data)
	if err != nil {
		logger.Debug(log)
		return
	}

	message := data["message"]
	delete(data, "message")
	if message == nil {
		return
	}

	level := logrus.DebugLevel
	_level, _ := data["level"].(float64)
	delete(data, "level")
	switch _level {
	case 200:
		level = logrus.InfoLevel
	case 300:
		level = logrus.WarnLevel
	case 400:
		level = logrus.ErrorLevel
	}

	logger.WithFields(data).Log(level, message)
}

func (e *Executor) logStream(c chan string, logger *logrus.Entry) {
	for log := range c {
		e.logLine(log, logger)
	}
}

func (e *Executor) Run(id string, data StdinData) (bool, error) {
	// ignoramos el error, no es posible que falle
	d, _ := json.Marshal(data)

	logger := e.parentLogger.WithFields(logrus.Fields{
		"sc_task.id":     id, // TODO se reemplaza por transaction.id, remover
		"transaction.id": id,
	})

	logger.
		WithField("event.type", []string{"start"}).
		// TODO: this fields are deprecated
		WithFields(logrus.Fields{
			"sc_task.cmd":     e.command,
			"sc_task.args":    e.args,
			"sc_task.action":  "started",
			"sc_task.started": 1,
		}).Info("task start")

	o := cmd.Options{
		Buffered:       false,
		Streaming:      true,
		LineBufferSize: 1 * 1000 * 1000, // 1 MB
	}
	c := cmd.NewCmdOptions(o, e.command, e.args...)
	go e.logStream(c.Stdout, logger)
	go e.logStream(c.Stderr, logger)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	var finalStatus cmd.Status

	for {
		select {
		case <-ctx.Done():
			c.Stop()
		case finalStatus = <-c.StartWithStdin(bytes.NewReader(d)):
			goto NEXT
		}
	}

NEXT:

	logger = logger.
		WithFields(logrus.Fields{
			"event.duration":    finalStatus.StopTs - finalStatus.StartTs,
			"event.type":        []string{"end"},
			"process.exit_code": finalStatus.Exit,
		}).
		// TODO: this fields are deprecated
		WithFields(logrus.Fields{
			"sc_task.duration": (finalStatus.StopTs - finalStatus.StartTs) / int64(time.Millisecond),
			"sc_task.action":   "finished",
			"sc_task.finished": 1,
		})

	ok := finalStatus.Exit == 0
	if ok {
		// todo bien
		logger = logger.
			WithField("event.outcome", "success").
			// TODO: this fields are deprecated
			WithFields(logrus.Fields{
				"sc_task.status":  "success",
				"sc_task.success": 1,
			})
		logger.Info("task finish")
	} else {
		// algo falló
		logger = logger.
			WithField("event.outcome", "failure").
			// TODO: this fields are deprecated
			WithFields(logrus.Fields{
				"sc_task.status": "fail",
				"sc_task.fail":   1,
			})

		if finalStatus.Error != nil {
			logger = logger.WithError(finalStatus.Error)
		}
		logger.Error("task finish")
	}

	return ok, nil
}

func NewExec(command string, args []string, timeout time.Duration, l *logrus.Entry) *Executor {
	return &Executor{
		command: command,
		args:    args,
		timeout: timeout,
		parentLogger: l.WithFields(logrus.Fields{
			"event.category":       []string{"process"},
			"event.dataset":        "subcommander.exec",
			"process.args":         args,
			"process.args_count":   len(args),
			"process.command_line": command,
		}),
	}
}
