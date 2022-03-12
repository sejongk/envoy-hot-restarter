package util

import (
	log "github.com/sirupsen/logrus"
)

var (
	logger log.FieldLogger
)

func init() {
	logger = log.New()
	log.SetLevel(log.DebugLevel)
}

func GetLogger() log.FieldLogger {
	return logger
}
