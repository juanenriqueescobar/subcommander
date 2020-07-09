package internal

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/juanenriqueescobar/cmd"
	"github.com/sirupsen/logrus"
)

type Executor struct {
	parentLogger *logrus.Entry
	command      string
	args         []string
}

func (e *Executor) logLine(log string, logger *logrus.Entry) {
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(log), &data)
	if err != nil {
		logger.WithError(err).Error(log)
		return
	}

	message := data["message"]
	delete(data, "message")
	if message == nil {
		return
	}

	level := logrus.DebugLevel
	_level, _ := data["level"].(float64)
	switch _level {
	case 200:
		level = logrus.InfoLevel
	case 300:
		level = logrus.WarnLevel
	case 400:
		level = logrus.ErrorLevel
	}
	delete(data, "level")
	data["_cmd"] = true
	logger.WithFields(data).Log(level, message)
}

func (e *Executor) logStream(c chan string, logger *logrus.Entry) {
	for log := range c {
		e.logLine(log, logger)
	}
}

func (e *Executor) run(id string, data string) bool {

	logger := e.parentLogger.WithField("_task_id", id)

	logger.WithFields(logrus.Fields{"_cmd": e.command, "_args": e.args}).Info("task start")
	o := cmd.Options{
		Streaming: true,
	}

	cmd := cmd.NewCmdOptions(o, e.command, e.args...)
	go e.logStream(cmd.Stdout, logger)
	statusChan := cmd.StartWithStdin(strings.NewReader(data))
	finalStatus := <-statusChan

	logger = logger.WithField("_duration", (finalStatus.StopTs-finalStatus.StartTs)/int64(time.Millisecond))

	// todo bien
	if finalStatus.Exit == 0 {
		logger.Info("task finish")
		return true
	}

	// algo falló
	if finalStatus.Error != nil {
		logger.WithError(finalStatus.Error).Error("task finish")
	} else {
		logger.Error("task finish")
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
