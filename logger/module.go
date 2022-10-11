package logger

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/rs/zerolog"
)

type LoggerModule struct {
	bus    modules.Bus
	logger modules.Logger
}

// Other loggers - main and ones injected in modules are branched out of mainLogger.
var mainLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()

// The idea is to create a logger for each module, so that we can easily filter logs by module.
// Keeping global logger too, because sometimes we need to log outside of modules.
func GlobalLogger() zerolog.Logger {

	// Should this be "module" = "none" instead?
	return mainLogger.With().Str("logger", "global").Logger()
}

var _ modules.LoggerModule = &LoggerModule{}

const (
	ModuleName = "logger"
)

func (lm *LoggerModule) CreateLoggerForModule(moduleName string) modules.Logger {
	return mainLogger.With().Str("module", moduleName).Logger()
}

func Create(configPath, genesisPath string) (modules.LoggerModule, error) {
	lm := new(LoggerModule)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	lm.InitLogger(configPath)

	c, err := lm.InitConfig(configPath)
	if err != nil {
		return nil, err
	}
	config := (c).(*LoggerConfig)

	zlLevel, err := zerolog.ParseLevel(config.GetLevel().String())
	if err != nil {
		return nil, err
	}
	zerolog.SetGlobalLevel(zlLevel)

	if config.GetFormat() == LogFormat_pretty {
		mainLogger = mainLogger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return lm, nil
}

func (lm *LoggerModule) InitConfig(pathToConfigJSON string) (config modules.IConfig, err error) {
	data, err := ioutil.ReadFile(pathToConfigJSON)
	if err != nil {
		return
	}
	// over arching configuration file
	rawJSON := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &rawJSON); err != nil {
		lm.logger.Fatal().Err(err).Str("path", pathToConfigJSON).Msg("an error occurred unmarshalling the config file")
	}
	// telemetry specific configuration file
	config = new(LoggerConfig)
	err = json.Unmarshal(rawJSON[lm.GetModuleName()], config)
	return
}

func (lm *LoggerModule) InitGenesis(pathToGenesisJSON string) (genesis modules.IGenesis, err error) {
	return // No-op
}

func (lm *LoggerModule) Start() error {
	return nil
}

func (lm *LoggerModule) Stop() error {
	return nil
}

func (lm *LoggerModule) GetModuleName() string {
	return ModuleName
}

func (lm *LoggerModule) SetBus(bus modules.Bus) {
	lm.bus = bus
}

func (lm *LoggerModule) GetBus() modules.Bus {
	if lm.bus == nil {
		lm.logger.Fatal().Msg("Bus is not initialized")
	}
	return lm.bus
}

func (lm *LoggerModule) InitLogger(pathToConfigJSON string) {
	lm.logger = lm.CreateLoggerForModule(lm.GetModuleName())
}

func (lm *LoggerModule) GetLogger() modules.Logger {
	return lm.logger
}
