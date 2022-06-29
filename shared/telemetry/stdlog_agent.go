package telemetry

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/pokt-network/pocket/shared/modules"
)

// INFO LogginLevel = iota
// ERROR
// DEBUG
// WARNING
// FATAL
// ALL

type StdLogAgent struct {
	*sync.Mutex
	*log.Logger
	level modules.LogLevel
}

func CreateStdLogAgent(level modules.LogLevel) modules.LogAgent {
	return NewLogger(level, os.Stdout)
}

func NewLogger(level modules.LogLevel, w io.Writer) *StdLogAgent {
	return &StdLogAgent{
		Mutex:  &sync.Mutex{},
		Logger: log.New(w, "[pocket]", 0),
		level:  level,
	}
}

func (sla *StdLogAgent) decorate(decor string, args []any) []any {
	fArgs := []any{decor}
	fArgs = append(fArgs, args...)
	return fArgs
}

func (sla *StdLogAgent) logIfAtLevel(level modules.LogLevel, namespace string, args ...any) {
	if sla.level != modules.LOG_LEVEL_NONE && (sla.level == level || sla.level == modules.LOG_LEVEL_ALL) {
		defer sla.Unlock()
		sla.Lock()

		namespacedArgs := []any{
			namespace,
		}
		namespacedArgs = append(namespacedArgs, args...)
		logLine := sla.decorate(fmt.Sprintf("[%s]:", level), namespacedArgs)

		if sla.level == modules.LOG_LEVEL_FATAL {
			sla.Logger.Fatal(logLine...)
		} else {
			sla.Logger.Println(logLine...)
		}
	}
}

// Interface for logging
func (sla *StdLogAgent) Info(namespace string, args ...any) {
	sla.logIfAtLevel(modules.LOG_LEVEL_INFO, namespace, args...)
}

func (sla *StdLogAgent) Error(namespace string, args ...any) {
	sla.logIfAtLevel(modules.LOG_LEVEL_ERROR, namespace, args...)
}

func (sla *StdLogAgent) Warn(namespace string, args ...any) {
	sla.logIfAtLevel(modules.LOG_LEVEL_WARN, namespace, args...)
}

func (sla *StdLogAgent) Debug(namespace string, args ...any) {
	sla.logIfAtLevel(modules.LOG_LEVEL_DEBUG, namespace, args...)
}

func (sla *StdLogAgent) Fatal(namespace string, args ...any) {
	sla.logIfAtLevel(modules.LOG_LEVEL_FATAL, namespace, args...)
}

func (sla *StdLogAgent) Log(namespace string, args ...any) {
	sla.logIfAtLevel(modules.LOG_LEVEL_ALL, namespace, args...)
}

func (sla *StdLogAgent) SetLevel(l modules.LogLevel) {
	sla.level = l
}
