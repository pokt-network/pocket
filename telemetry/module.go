package telemetry

import (
	"encoding/json"
	"github.com/pokt-network/pocket/shared/modules"
	"io/ioutil"
	"log"
)

var _ modules.Module = &telemetryModule{}
var _ modules.TelemetryConfig = &TelemetryConfig{}

const (
	TelemetryModuleName = "telemetry"
)

// TODO(pocket/issues/99): Add a switch statement and configuration variable when support for other telemetry modules is added.
func Create(configPath, genesisPath string) (modules.TelemetryModule, error) {
	tm := new(telemetryModule)
	c, err := tm.InitConfig(configPath)
	if err != nil {
		return nil, err
	}
	cfg := c.(*TelemetryConfig)
	if cfg.GetEnabled() {
		return CreatePrometheusTelemetryModule(cfg)
	} else {
		return CreateNoopTelemetryModule(cfg)
	}
}

type telemetryModule struct{}

func (t *telemetryModule) InitConfig(pathToConfigJSON string) (config modules.ConfigI, err error) {
	data, err := ioutil.ReadFile(pathToConfigJSON)
	if err != nil {
		return
	}
	// over arching configuration file
	rawJSON := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &rawJSON); err != nil {
		log.Fatalf("[ERROR] an error occurred unmarshalling the config.json file: %v", err.Error())
	}
	// persistence specific configuration file
	config = new(TelemetryConfig)
	err = json.Unmarshal(rawJSON[t.GetModuleName()], config)
	return
}

func (t *telemetryModule) GetModuleName() string                                      { return TelemetryModuleName }
func (t *telemetryModule) InitGenesis(_ string) (genesis modules.GenesisI, err error) { return }
func (t *telemetryModule) SetBus(bus modules.Bus)                                     {}
func (t *telemetryModule) GetBus() modules.Bus                                        { return nil }
func (t *telemetryModule) Start() error                                               { return nil }
func (t *telemetryModule) Stop() error                                                { return nil }
