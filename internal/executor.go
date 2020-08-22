package internal

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/juanenriqueescobar/cmd"
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
}

func (e *Executor) logLine(log string, logger *logrus.Entry) {

	log = strings.TrimSpace(log)
	if len(log) == 0 {
		return
	}

	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(log), &data)
	if err != nil {
		logger.WithField("_jsonlog", false).Debug(log)
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

	data["_jsonlog"] = true
	logger.WithFields(data).Log(level, message)
}

func (e *Executor) logStream(c chan string, logger *logrus.Entry) {
	for log := range c {
		e.logLine(log, logger)
	}
}

// Deprecated
func (e *Executor) run(id string, data StdinData) (bool, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(data)
	if err != nil {
		return false, err
	}
	return e.runReader(id, buf), nil
}

func (e *Executor) runReader(id string, data io.Reader) bool {

	logger := e.parentLogger.WithFields(logrus.Fields{
		"trace.id":   id, // TODO esto no es el trace id, falta recibirlo
		"sc_task.id": id,
	})

	logger.WithFields(logrus.Fields{
		"sc_task": logrus.Fields{
			"cmd":     e.command,
			"args":    e.args,
			"action":  "started",
			"started": 1,
		},
	}).Info("task start")
	o := cmd.Options{
		Streaming: true,
	}

	cmd := cmd.NewCmdOptions(o, e.command, e.args...)
	go e.logStream(cmd.Stdout, logger)

	statusChan := cmd.StartWithStdin(data)
	finalStatus := <-statusChan

	fields := logrus.Fields{
		"sc_task": logrus.Fields{
			"duration": (finalStatus.StopTs - finalStatus.StartTs) / int64(time.Millisecond),
			"action":   "finished",
			"finished": 1,
		},
	}
	// todo bien
	if finalStatus.Exit == 0 {
		fields["sc_task.status"] = "success"
		fields["sc_task.success"] = 1
		logger.WithFields(fields).Info("task finish")
		return true
	}

	fields["sc_task.status"] = "fail"
	fields["sc_task.fail"] = 1
	// algo falló
	if finalStatus.Error != nil {
		logger.WithFields(fields).WithError(finalStatus.Error).Error("task finish")
	} else {
		logger.WithFields(fields).Error("task finish")
	}
	return false
}

func NewExec(command string, args []string, l *logrus.Entry) *Executor {
	return &Executor{
		command:      command,
		args:         args,
		parentLogger: l,
	}
}
