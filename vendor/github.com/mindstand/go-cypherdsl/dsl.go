package go_cypherdsl

import (
	"errors"
	"github.com/sirupsen/logrus"
)

var externalLog *logrus.Entry

var log = getLogger()

func getLogger() *logrus.Entry {
	if externalLog == nil {
		//create default logger
		toReturn := logrus.New()

		return toReturn.WithField("source", "go-cypherdsl")
	}

	return externalLog
}

func SetLogger(logger *logrus.Entry) error {
	if logger == nil {
		return errors.New("logger can not be nil")
	}
	externalLog = logger
	return nil
}
