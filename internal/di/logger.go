package di

import (
	"github.com/juanenriqueescobar/subcommander/internal/ec2"
	"github.com/sirupsen/logrus"
)

func logger2(i ec2.Info, logger *logrus.Logger) *logrus.Entry {
	if i.ID() != "" {
		return logger.WithField("_instance", i.ID())
	}

	return logrus.NewEntry(logger)
}
