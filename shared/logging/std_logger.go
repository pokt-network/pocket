package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type stdLogger struct {
	*log.Logger
	*sync.Mutex
	namespace Namespace
	prefix    string
	level     LogLevel
}

// DISCUSS(team): The only reason I've made this function variadic for the 'out' arg is testing
// I don't know how to utilize mockgen to mock standard library, specificaly (os.Stdout).
// Thus, if you find another solution, do away with the variadic signature and only keep the 'level' argument.
func CreateStdLogger(level LogLevel, namespace Namespace, prefix string, out ...io.Writer) Logger {
	var logger *log.Logger

	if len(out) > 0 {
		logger = log.New(out[0], "", log.LstdFlags)
	} else {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	return &stdLogger{
		Mutex:     &sync.Mutex{},
		Logger:    logger,
		level:     LOG_LEVEL_ALL,
		namespace: namespace,
		prefix:    prefix,
	}
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
			fmt.Sprintf("[%s]", sla.namespace),
		}
		namespacedArgs = append(namespacedArgs, args...)
		logLine := sla.decorate(fmt.Sprintf("| [%s][%s]:", sla.prefix, level), namespacedArgs)

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
