package di

import (
	"github.com/sirupsen/logrus"
)

func logger2(logger *logrus.Logger) *logrus.Entry {
	return logrus.NewEntry(logger)
}
