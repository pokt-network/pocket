package logger

import (
	"os"
	"strings"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/rs/zerolog"
)

type loggerModule struct {
	bus    modules.Bus
	logger modules.Logger
	config modules.LoggerConfig
}

// All loggers branch out of mainLogger, that way configuration changes to mainLogger propagate to others.
var mainLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()

// The idea is to create a logger for each module, so that we can easily filter logs by module.
// But we also need a global logger, because sometimes we need to log outside of modules, e.g. when the process just
// started, and modules are not initiated yet.
var Global = new(loggerModule).CreateLoggerForModule("global")

var _ modules.LoggerModule = &loggerModule{}

const (
	ModuleName = "logger"
)

var pocketLogLevelToZeroLog = map[LogLevel]zerolog.Level{
	LogLevel_LOG_LEVEL_UNSPECIFIED: zerolog.NoLevel,
	LogLevel_LOG_LEVEL_DEBUG:       zerolog.DebugLevel,
	LogLevel_LOG_LEVEL_INFO:        zerolog.InfoLevel,
	LogLevel_LOG_LEVEL_WARN:        zerolog.WarnLevel,
	LogLevel_LOG_LEVEL_ERROR:       zerolog.ErrorLevel,
	LogLevel_LOG_LEVEL_FATAL:       zerolog.FatalLevel,
	LogLevel_LOG_LEVEL_PANIC:       zerolog.PanicLevel,
}

var pocketLogFormatToEnum = map[string]LogFormat{
	"json":   LogFormat_LOG_FORMAT_JSON,
	"pretty": LogFormat_LOG_FORMAT_PRETTY,
}

func Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	return new(loggerModule).Create(runtimeMgr)
}

func (*loggerModule) CreateLoggerForModule(moduleName string) modules.Logger {
	return mainLogger.With().Str("module", moduleName).Logger()
}

func (*loggerModule) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	cfg := runtimeMgr.GetConfig()
	m := loggerModule{
		config: cfg.GetLoggerConfig(),
	}

	m.InitLogger()

	// Mapping config string value to the proto enum
	if pocketLogLevel, ok := LogLevel_value[`LogLevel_LOG_LEVEL_`+strings.ToUpper(m.config.GetLevel())]; ok {
		zerolog.SetGlobalLevel(pocketLogLevelToZeroLog[LogLevel(pocketLogLevel)])
	} else {
		zerolog.SetGlobalLevel(zerolog.NoLevel)
	}

	if pocketLogFormatToEnum[m.config.GetFormat()] == LogFormat_LOG_FORMAT_PRETTY {
		mainLogger = mainLogger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		mainLogger.Info().Msg("using pretty log format")
	}

	return &m, nil
}

func (m *loggerModule) Start() error {
	return nil
}

func (m *loggerModule) Stop() error {
	return nil
}

func (m *loggerModule) GetModuleName() string {
	return ModuleName
}

func (m *loggerModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *loggerModule) GetBus() modules.Bus {
	if m.bus == nil {
		m.logger.Fatal().Msg("Bus is not initialized")
	}
	return m.bus
}

func (m *loggerModule) InitLogger() {
	m.logger = m.CreateLoggerForModule(m.GetModuleName())
}

func (m *loggerModule) GetLogger() modules.Logger {
	return m.logger
}
