package types

// TODO(team): Consolidate the logging library with the rest of the codebase

import (
	"io"
	log "log"
	sync "sync"
)

type (
	logger struct {
		sync.Mutex
		*log.Logger
		suppressed bool
	}

	Logger interface {
		Debug(...interface{})
		Log(...interface{})
		Info(...interface{})
		Error(...interface{})
		Warn(...interface{})
		Suppress(bool)
	}
)

func NewLogger(w io.Writer) *logger {
	return &logger{
		Logger:     log.New(w, "[pocket]", 0),
		suppressed: false,
	}
}

func (l *logger) decorate(decor string, args []interface{}) []interface{} {
	fArgs := []interface{}{decor}
	fArgs = append(fArgs, args...)
	return fArgs
}

func (l *logger) Debug(args ...interface{}) {
	defer l.Unlock()
	l.Lock()
	if !l.suppressed {
		l.Logger.Println(l.decorate("[DEBUG]:", args)...)
	}
}

func (l *logger) Log(args ...interface{}) {
	defer l.Unlock()
	l.Lock()
	if !l.suppressed {
		l.Logger.Println(l.decorate("[LOG]:", args)...)
	}
}

func (l *logger) Info(args ...interface{}) {
	defer l.Unlock()
	l.Lock()
	if !l.suppressed {
		l.Logger.Println(l.decorate("[INFO]:", args)...)
	}
}

func (l *logger) Error(args ...interface{}) {
	defer l.Unlock()
	l.Lock()
	if !l.suppressed {
		l.Logger.Println(l.decorate("[ERROR]:", args)...)
	}
}

func (l *logger) Warn(args ...interface{}) {
	defer l.Unlock()
	l.Lock()
	if !l.suppressed {
		l.Logger.Println(l.decorate("[WARNING]:", args)...)
	}
}

func (l *logger) Suppress(suppress bool) {
	defer l.Unlock()
	l.Lock()
	l.suppressed = suppress
}
