package types

import (
	log "log"
)

type (
	logger struct{}

	Logger interface {
		Debug(...interface{})
		Log(...interface{})
		Info(...interface{})
		Error(...interface{})
		Warn(...interface{})
	}
)

func NewLogger() *logger {
	return &logger{}
}

func (l *logger) Debug(args ...interface{}) {
	fArgs := []interface{}{"[DEBUG]"}
	fArgs = append(fArgs, args...)
	log.Println(fArgs...)
}
func (l *logger) Log(args ...interface{}) {
	fArgs := []interface{}{"[LOG]"}
	fArgs = append(fArgs, args...)
	log.Println(fArgs...)
}

func (l *logger) Info(args ...interface{}) {
	fArgs := []interface{}{"[INFO]"}
	fArgs = append(fArgs, args...)
	log.Println(fArgs...)
}
func (l *logger) Error(args ...interface{}) {
	fArgs := []interface{}{"[Error]"}
	fArgs = append(fArgs, args...)
	log.Println(fArgs...)
}

func (l *logger) Warn(args ...interface{}) {
	fArgs := []interface{}{"[WARNING]"}
	fArgs = append(fArgs, args...)
	log.Println(fArgs...)
}
