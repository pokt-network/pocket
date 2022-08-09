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
