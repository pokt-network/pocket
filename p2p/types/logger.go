package types

import (
	"io"
	log "log"
)

type (
	logger struct {
		*log.Logger
	}

	Logger interface {
		Debug(...interface{})
		Log(...interface{})
		Info(...interface{})
		Error(...interface{})
		Warn(...interface{})
	}
)

func NewLogger(w io.Writer) *logger {
	return &logger{
		Logger: log.New(w, "[pocket]", 0),
	}
}

func (l *logger) decorate(decor string, args []interface{}) []interface{} {
	fArgs := []interface{}{decor}
	fArgs = append(fArgs, args...)
	return fArgs
}

func (l *logger) Debug(args ...interface{}) {
	l.Logger.Println(l.decorate("[DEBUG]:", args)...)
}

func (l *logger) Log(args ...interface{}) {
	l.Logger.Println(l.decorate("[LOG]:", args)...)
}

func (l *logger) Info(args ...interface{}) {
	l.Logger.Println(l.decorate("[INFO]:", args)...)
}

func (l *logger) Error(args ...interface{}) {
	l.Logger.Println(l.decorate("[ERROR]:", args)...)
}

func (l *logger) Warn(args ...interface{}) {
	l.Logger.Println(l.decorate("[WARNING]:", args)...)
}
