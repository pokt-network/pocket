package logging

var (
	singletonLogger = CreateStdLogger(LOG_LEVEL_ALL)
)

func GetGlobalLogger() Logger {
	return singletonLogger
}
