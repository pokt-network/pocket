package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/rs/zerolog"
)

type loggerModule struct {
	zerolog.Logger
	bus    modules.Bus
	config modules.LoggerConfig
}

// The idea is to create a logger for each module, so that we can easily filter logs by module.
// But we also need a global logger, because sometimes we need to log outside of modules, e.g. when the process just
// started, and modules are not initiated yet.
var Global = loggerModule{
	// All loggers branch out of mainLogger, that way configuration changes to mainLogger propagate to others.
	Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
}

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
	return Global.Logger.With().Str("module", moduleName).Logger()
}

func (*loggerModule) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	cfg := runtimeMgr.GetConfig()

	Global.config = cfg.GetLoggerConfig()
	Global.InitLogger()

	// Mapping config string value to the proto enum
	if pocketLogLevel, ok := LogLevel_value[`LOG_LEVEL_`+strings.ToUpper(Global.config.GetLevel())]; ok {
		zerolog.SetGlobalLevel(pocketLogLevelToZeroLog[LogLevel(pocketLogLevel)])
	} else {
		zerolog.SetGlobalLevel(zerolog.NoLevel)
	}

	if pocketLogFormatToEnum[Global.config.GetFormat()] == LogFormat_LOG_FORMAT_PRETTY {
		logStructure := zerolog.ConsoleWriter{Out: os.Stdout}
		logStructure.FormatLevel = func(i interface{}) string {
			return fmt.Sprintf("level=%s", strings.ToUpper(i.(string)))
		}

		Global.Logger = Global.Logger.Output(logStructure)
		Global.Logger.Info().Msg("using pretty log format")
	}

	return &Global, nil
}

func (m *loggerModule) Start() error {
	Global.Logger = m.CreateLoggerForModule("global")
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
		m.Logger.Fatal().Msg("Bus is not initialized")
	}
	return m.bus
}

func (m *loggerModule) InitLogger() {
	m.Logger = m.CreateLoggerForModule(m.GetModuleName())
}

func (m *loggerModule) GetLogger() modules.Logger {
	return m.Logger
}

// SetFields sets the fields for the global logger
func (m *loggerModule) SetFields(fields map[string]interface{}) {
	m.Logger = m.Logger.With().Fields(fields).Logger()
}
