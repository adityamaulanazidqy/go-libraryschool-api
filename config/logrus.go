package config

import (
	"github.com/sirupsen/logrus"
	"os"
)

func LogrusLogger() *logrus.Logger {
	var log = logrus.New()
	log.Out = os.Stdout
	log.Level = logrus.DebugLevel
	return log
}
