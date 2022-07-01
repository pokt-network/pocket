package logging

import "strings"

type LogLevel string

const (
	LOG_LEVEL_NONE  LogLevel = "NONE"
	LOG_LEVEL_INFO           = "INFO"
	LOG_LEVEL_ERROR          = "ERROR"
	LOG_LEVEL_DEBUG          = "DEBUG"
	LOG_LEVEL_WARN           = "WARN"
	LOG_LEVEL_FATAL          = "FATAL"
	LOG_LEVEL_ALL            = "LOG"
)

type Namespace string

const (
	CONSENSUS_NAMESPACE   Namespace = "CONSENSUS"
	P2P_NAMESPACE                   = "P2P"
	UTILITY_NAMESPACE               = "UTILITY"
	PERSISTENCE_NAMESPACE           = "PERSISTENCE"
	GLOBAL_NAMESPACE                = "GLOBAL"
)

// Interface for logging
type Logger interface {
	SetLevel(LogLevel)
	SetNamespace(Namespace)

	Info(args ...any)  // level = info
	Error(args ...any) // level = error
	Warn(args ...any)  // level = warn
	Debug(args ...any) // level = debug
	Fatal(args ...any) // level = fatal
	Log(args ...any)   // level = all
}

type LoggerConfig map[Namespace]LogLevel

func GetLevel(level string) LogLevel {
	switch level {
	case strings.ToLower(string(LOG_LEVEL_ALL)):
		return LOG_LEVEL_ALL
	case strings.ToLower(string(LOG_LEVEL_FATAL)):
		return LOG_LEVEL_FATAL
	case strings.ToLower(string(LOG_LEVEL_ERROR)):
		return LOG_LEVEL_ERROR
	case strings.ToLower(string(LOG_LEVEL_DEBUG)):
		return LOG_LEVEL_DEBUG
	case strings.ToLower(string(LOG_LEVEL_WARN)):
		return LOG_LEVEL_WARN
	case strings.ToLower(string(LOG_LEVEL_INFO)):
		return LOG_LEVEL_INFO
	default:
		return LOG_LEVEL_NONE
	}
}
