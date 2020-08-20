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

func (e *Executor) run(id string, data StdinData) (bool, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(data)
	if err != nil {
		return false, err
	}
	return e.runReader(id, buf), nil
}

func (e *Executor) runReader(id string, data io.Reader) bool {

	logger := e.parentLogger.WithField("_task_id", id)

	logger.WithFields(logrus.Fields{"_cmd": e.command, "_args": e.args}).Info("task start")
	o := cmd.Options{
		Streaming: true,
	}

	cmd := cmd.NewCmdOptions(o, e.command, e.args...)
	go e.logStream(cmd.Stdout, logger)

	statusChan := cmd.StartWithStdin(data)
	finalStatus := <-statusChan

	logger = logger.WithField("_duration", (finalStatus.StopTs-finalStatus.StartTs)/int64(time.Millisecond))

	// todo bien
	if finalStatus.Exit == 0 {
		logger.WithField("_task_ok", 1).Info("task finish")
		return true
	}

	// algo falló
	if finalStatus.Error != nil {
		logger.WithField("_task_fail", 1).WithError(finalStatus.Error).Error("task finish")
	} else {
		logger.WithField("_task_fail", 1).Error("task finish")
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
