package logging

import "sync"

var singletonLogger Logger
var singletonLock = &sync.Mutex{}

func GetGlobalLogger() Logger {
	if singletonLogger == nil {
		singletonLock.Lock()
		defer singletonLock.Unlock()
		singletonLogger = CreateStdLogger(LOG_LEVEL_ALL, GLOBAL_NAMESPACE, POCKET_LOGS_PREFIX)
	}
	return singletonLogger
}

func Info(args ...any) {
	GetGlobalLogger().Info(args...)
}

func Error(args ...any) {
	GetGlobalLogger().Error(args...)
}

func Warn(args ...any) {
	GetGlobalLogger().Warn(args...)
}

func Debug(args ...any) {
	GetGlobalLogger().Debug(args...)
}

func Fatal(args ...any) {
	GetGlobalLogger().Fatal(args...)
}

func Log(args ...any) {
	GetGlobalLogger().Log(args...)
}
