package types

import sync "sync"

type (
	// TODO(team): replace with whatever logged decided upon when telemetry hits the code base
	logger struct { // a temporary logger struct to allow flexible injection of log functions
		sync.Mutex
		print func(...interface{}) (int, error)
	}

	Logger interface {
		Debug(...interface{})
		Log(...interface{})
		Info(...interface{})
		Error(...interface{})
		Warn(...interface{})
	}
)

func (l *logger) Debug(...interface{}) {}
func (l *logger) Log(...interface{})   {}
func (l *logger) Info(...interface{})  {}
func (l *logger) Error(...interface{}) {}
func (l *logger) Warn(...interface{})  {}
