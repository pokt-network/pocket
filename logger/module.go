package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/rs/zerolog"
)

var _ modules.Module = &loggerModule{}

type loggerModule struct {
	zerolog.Logger
	bus    modules.Bus
	config *configs.LoggerConfig
}

// Each module should have it's own logger to easily configure & filter logs by module.

// A Global logger is also created to enable logging outside of modules (e.g. when the node is starting).
// All loggers branch out of Global, that way configuration changes to Global propagate to others.
var Global loggerModule

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

// init is called when the package is imported.
// It is used to initialize the global logger.
func init() {
	Global = loggerModule{
		Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}
}

func Create(bus modules.Bus) (modules.Module, error) {
	return new(loggerModule).Create(bus)
}

func (*loggerModule) CreateLoggerForModule(moduleName string) modules.Logger {
	return Global.Logger.With().Str("module", moduleName).Logger()
}

func (*loggerModule) Create(bus modules.Bus) (modules.Module, error) {
	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()
	m := &loggerModule{
		config: cfg.Logger,
	}
	if err := bus.RegisterModule(m); err != nil {
		return nil, err
	}

	Global.config = m.config
	Global.CreateLoggerForModule("global")

	// Mapping config string value to the proto enum
	if pocketLogLevel, ok := configs.LogLevel_value[`LOG_LEVEL_`+strings.ToUpper(Global.config.GetLevel())]; ok {
		zerolog.SetGlobalLevel(pocketLogLevelToZeroLog[configs.LogLevel(pocketLogLevel)])
	} else {
		zerolog.SetGlobalLevel(zerolog.NoLevel)
	}

	if pocketLogFormatToEnum[Global.config.GetFormat()] == configs.LogFormat_LOG_FORMAT_PRETTY {
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
	return modules.LoggerModuleName
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

func (m *loggerModule) GetLogger() modules.Logger {
	return m.Logger
}

// INVESTIGATE(#420): https://github.com/pokt-network/pocket/issues/480
// SetFields sets the fields for the global logger
func (m *loggerModule) SetFields(fields map[string]any) {
	m.Logger = m.Logger.With().Fields(fields).Logger()
}

// UpdateFields updates the fields for the global logger
func (m *loggerModule) UpdateFields(fields map[string]any) {
	m.Logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		for k, v := range fields {
			c = c.Interface(k, v)
		}
		return c
	})
}
