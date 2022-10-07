package logger

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LoggerModule struct {
	bus          modules.Bus
	GlobalLogger zerolog.Logger
	ModuleLogger zerolog.Logger
}

var _ modules.LoggerModule = &LoggerModule{}

const (
	LoggerModuleName = "logger"
)

func (l *LoggerModule) GetLoggerForModule(moduleName string) zerolog.Logger {
	return l.GlobalLogger.With().Str("module", moduleName).Logger()
}

// curiosity question: why do we need genesisPath for modules that have no interest in it?
func Create(configPath, genesisPath string) (modules.LoggerModule, error) {
	lm := new(LoggerModule)
	lm.GlobalLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	lm.ModuleLogger = lm.GlobalLogger.With().Str("module", LoggerModuleName).Logger()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix // before or after init?

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
		lm.GlobalLogger = lm.GlobalLogger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return lm, nil
}

func (l *LoggerModule) InitConfig(pathToConfigJSON string) (config modules.IConfig, err error) {
	data, err := ioutil.ReadFile(pathToConfigJSON)
	if err != nil {
		return
	}
	// over arching configuration file
	rawJSON := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &rawJSON); err != nil {
		log.Fatal().Err(err).Str("path", pathToConfigJSON).Msg("an error occurred unmarshalling the config file")
	}
	// telemetry specific configuration file
	config = new(LoggerConfig)
	err = json.Unmarshal(rawJSON[l.GetModuleName()], config)
	return
}

func (l *LoggerModule) InitGenesis(pathToGenesisJSON string) (genesis modules.IGenesis, err error) {
	return // No-op
}

func (l *LoggerModule) Start() error {
	return nil
}

func (l *LoggerModule) Stop() error {
	return nil
}

func (l *LoggerModule) GetModuleName() string {
	return LoggerModuleName
}

func (l *LoggerModule) SetBus(bus modules.Bus) {
	l.bus = bus
}

func (l *LoggerModule) GetBus() modules.Bus {
	if l.bus == nil {
		l.ModuleLogger.Fatal().Msg("Bus is not initialized")
	}
	return l.bus
}
