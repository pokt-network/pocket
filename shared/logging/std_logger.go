package logging

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type stdLogger struct {
	*log.Logger
	*sync.Mutex
	namespace Namespace
	level     LogLevel
}

func CreateStdLogger(level LogLevel) Logger {
	return &stdLogger{
		Mutex:     &sync.Mutex{},
		Logger:    log.New(os.Stdout, "[pocket]", 0),
		level:     LOG_LEVEL_ALL,
		namespace: GLOBAL_NAMESPACE,
	}, nil
}

func (sla *stdLogger) decorate(decor string, args []any) []any {
	fArgs := []any{decor}
	fArgs = append(fArgs, args...)
	return fArgs
}

func (sla *stdLogger) logIfAtLevel(level LogLevel, args ...any) {
	if sla.level != LOG_LEVEL_NONE && (sla.level == level || sla.level == LOG_LEVEL_ALL) {
		defer sla.Unlock()
		sla.Lock()

		namespacedArgs := []any{
			sla.namespace,
		}
		namespacedArgs = append(namespacedArgs, args...)
		logLine := sla.decorate(fmt.Sprintf("[%s]:", level), namespacedArgs)

		if sla.level == LOG_LEVEL_FATAL {
			sla.Logger.Fatal(logLine...)
		} else {
			sla.Logger.Println(logLine...)
		}
	}
}

// Interface for logging
func (sla *stdLogger) Info(args ...any) {
	sla.logIfAtLevel(LOG_LEVEL_INFO, args...)
}

func (sla *stdLogger) Error(args ...any) {
	sla.logIfAtLevel(LOG_LEVEL_ERROR, args...)
}

func (sla *stdLogger) Warn(args ...any) {
	sla.logIfAtLevel(LOG_LEVEL_WARN, args...)
}

func (sla *stdLogger) Debug(args ...any) {
	sla.logIfAtLevel(LOG_LEVEL_DEBUG, args...)
}

func (sla *stdLogger) Fatal(args ...any) {
	sla.logIfAtLevel(LOG_LEVEL_FATAL, args...)
}

func (sla *stdLogger) Log(args ...any) {
	sla.logIfAtLevel(LOG_LEVEL_ALL, args...)
}

func (sla *stdLogger) SetLevel(l LogLevel) {
	sla.level = l
}

func (sla *stdLogger) SetNamespace(namespace Namespace) {
	sla.namespace = namespace
}
