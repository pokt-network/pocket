package logger

import (
	"os"
	"strings"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/rs/zerolog"
)

type loggerModule struct {
	bus    modules.Bus
	logger modules.Logger
	config *configs.LoggerConfig
}

// All loggers branch out of mainLogger, that way configuration changes to mainLogger propagate to others.
var mainLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()

// The idea is to create a logger for each module, so that we can easily filter logs by module.
// But we also need a global logger, because sometimes we need to log outside of modules, e.g. when the process just
// started, and modules are not initiated yet.
var Global = new(loggerModule).CreateLoggerForModule("global")

var _ modules.LoggerModule = &loggerModule{}

var pocketLogLevelToZeroLog = map[configs.LogLevel]zerolog.Level{
	configs.LogLevel_LOG_LEVEL_UNSPECIFIED: zerolog.NoLevel,
	configs.LogLevel_LOG_LEVEL_DEBUG:       zerolog.DebugLevel,
	configs.LogLevel_LOG_LEVEL_INFO:        zerolog.InfoLevel,
	configs.LogLevel_LOG_LEVEL_WARN:        zerolog.WarnLevel,
	configs.LogLevel_LOG_LEVEL_ERROR:       zerolog.ErrorLevel,
	configs.LogLevel_LOG_LEVEL_FATAL:       zerolog.FatalLevel,
	configs.LogLevel_LOG_LEVEL_PANIC:       zerolog.PanicLevel,
}

var pocketLogFormatToEnum = map[string]configs.LogFormat{
	"json":   configs.LogFormat_LOG_FORMAT_JSON,
	"pretty": configs.LogFormat_LOG_FORMAT_PRETTY,
}

func Create(bus modules.Bus) (modules.Module, error) {
	return new(loggerModule).Create(bus)
}

func (*loggerModule) CreateLoggerForModule(moduleName string) modules.Logger {
	return mainLogger.With().Str("module", moduleName).Logger()
}

func (*loggerModule) Create(bus modules.Bus) (modules.Module, error) {
	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()
	m := &loggerModule{
		config: cfg.Logger,
	}
	bus.GetModulesRegistry().RegisterModule(m)

	m.InitLogger()

	// Mapping config string value to the proto enum
	if pocketLogLevel, ok := configs.LogLevel_value[`LogLevel_LOG_LEVEL_`+strings.ToUpper(m.config.Level)]; ok {
		zerolog.SetGlobalLevel(pocketLogLevelToZeroLog[configs.LogLevel(pocketLogLevel)])
	} else {
		zerolog.SetGlobalLevel(zerolog.NoLevel)
	}

	if pocketLogFormatToEnum[m.config.Format] == configs.LogFormat_LOG_FORMAT_PRETTY {
		mainLogger = mainLogger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		mainLogger.Info().Msg("using pretty log format")
	}

	return m, nil
}

func (m *loggerModule) Start() error {
	return nil
}

func (m *loggerModule) Stop() error {
	return nil
}

func (m *loggerModule) GetModuleName() string {
	return modules.LoggerModuleName
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
