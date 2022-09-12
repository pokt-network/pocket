package telemetry

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/pokt-network/pocket/shared/modules"
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

func (t *telemetryModule) InitConfig(pathToConfigJSON string) (config modules.IConfig, err error) {
	data, err := ioutil.ReadFile(pathToConfigJSON)
	if err != nil {
		return
	}
	// over arching configuration file
	rawJSON := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &rawJSON); err != nil {
		log.Fatalf("[ERROR] an error occurred unmarshalling the %s file: %v", pathToConfigJSON, err.Error())
	}
	// telemetry specific configuration file
	config = new(TelemetryConfig)
	err = json.Unmarshal(rawJSON[t.GetModuleName()], config)
	return
}

func (t *telemetryModule) GetModuleName() string                                      { return TelemetryModuleName }
func (t *telemetryModule) InitGenesis(_ string) (genesis modules.IGenesis, err error) { return }
func (t *telemetryModule) SetBus(bus modules.Bus)                                     {}
func (t *telemetryModule) GetBus() modules.Bus                                        { return nil }
func (t *telemetryModule) Start() error                                               { return nil }
func (t *telemetryModule) Stop() error                                                { return nil }
