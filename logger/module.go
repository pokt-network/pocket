package logger

import (
	"os"

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

	zlLevel, err := zerolog.ParseLevel(m.config.GetLevel())
	if err != nil {
		m.logger.Err(err).Msg("couldn't parse log level")
		return nil, err
	}
	zerolog.SetGlobalLevel(zlLevel)

	if m.config.GetFormat() == LogFormat_pretty.String() {
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
