//+build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/juanenriqueescobar/subcommander/internal/commander"
	"github.com/sirupsen/logrus"
)

func Logger() *logrus.Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "log.level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "function.name", // non-ECS
			// logrus.FieldKeyLogrusError: "error.message",
		},
	})
	logrus.ErrorKey = "error.message" // TODO why?
	logger.SetLevel(logrus.DebugLevel)
	return logger
}

func Commander(logger *logrus.Logger, args []string) (*commander.Commander, error) {
	wire.Build(stdset)
	return &commander.Commander{}, nil
}
