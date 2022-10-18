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

// Other loggers - main and ones injected in modules are branched out of mainLogger.
var mainLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()

// The idea is to create a logger for each module, so that we can easily filter logs by module.
// Keeping global logger too, because sometimes we need to log outside of modules.
func GlobalLogger() modules.Logger {

	// Should this be "module" = "none" instead?
	return mainLogger.With().Str("logger", "global").Logger()
}

var _ modules.LoggerModule = &loggerModule{}

const (
	// Should this be a "log" module instead, so we can have "log" in the config file? (end envar too: e.g. POCKET_LOG_LEVEL)
	ModuleName = "logger"
)

func Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	return new(loggerModule).Create(runtimeMgr)
}

func (*loggerModule) CreateLoggerForModule(moduleName string) modules.Logger {
	return mainLogger.With().Str("module", moduleName).Logger()
}

func (*loggerModule) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	var m *loggerModule

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	m.InitLogger()

	cfg := runtimeMgr.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		m.logger.Err(err).Msg("config validation failed")
		return nil, err
	}

	m.config = cfg.GetLoggerConfig()

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

	return m, nil
}

// func (lm *loggerModule) InitConfig(pathToConfigJSON string) (config modules.IConfig, err error) {
// 	data, err := ioutil.ReadFile(pathToConfigJSON)
// 	if err != nil {
// 		return
// 	}
// 	// over arching configuration file
// 	rawJSON := make(map[string]json.RawMessage)
// 	if err = json.Unmarshal(data, &rawJSON); err != nil {
// 		lm.logger.Fatal().Err(err).Str("path", pathToConfigJSON).Msg("an error occurred unmarshalling the config file")
// 	}

// 	loggerConfig := &LoggerConfig{}

// 	// Protojson is necessary to unmarshal the enum values
// 	if err = protojson.Unmarshal(rawJSON[lm.GetModuleName()], loggerConfig); err != nil {
// 		lm.logger.Fatal().Err(err).Str("path", pathToConfigJSON).Msg("can't unmarshal logger config")
// 	}

// 	return loggerConfig, err
// }

func (lm *loggerModule) Start() error {
	return nil
}

func (lm *loggerModule) Stop() error {
	return nil
}

func (lm *loggerModule) GetModuleName() string {
	return ModuleName
}

func (lm *loggerModule) SetBus(bus modules.Bus) {
	lm.bus = bus
}

func (lm *loggerModule) GetBus() modules.Bus {
	if lm.bus == nil {
		lm.logger.Fatal().Msg("Bus is not initialized")
	}
	return lm.bus
}

func (*loggerModule) ValidateConfig(cfg modules.Config) error {
	return nil
}

func (lm *loggerModule) InitLogger() {
	lm.logger = lm.CreateLoggerForModule(lm.GetModuleName())
}

func (lm *loggerModule) GetLogger() modules.Logger {
	return lm.logger
}
