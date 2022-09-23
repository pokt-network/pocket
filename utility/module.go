package utility

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/pokt-network/pocket/utility/types"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.UtilityModule = &UtilityModule{}
var _ modules.UtilityConfig = &types.UtilityConfig{}

type UtilityModule struct {
	bus     modules.Bus
	Mempool types.Mempool
}

const (
	UtilityModuleName = "utility"
)

func Create(configPath, genesisPath string) (modules.UtilityModule, error) {
	u := new(UtilityModule)
	c, err := u.InitConfig(configPath)
	if err != nil {
		return nil, err
	}
	config := (c).(*types.UtilityConfig)
	return &UtilityModule{
		Mempool: types.NewMempool(config.Max_Mempool_Transaction_Bytes, config.Max_Mempool_Transactions),
	}, nil
}

func (u *UtilityModule) InitConfig(pathToConfigJSON string) (config modules.IConfig, err error) {
	data, err := ioutil.ReadFile(pathToConfigJSON)
	if err != nil {
		return
	}
	// over arching configuration file
	rawJSON := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &rawJSON); err != nil {
		log.Fatalf("[ERROR] an error occurred unmarshalling the %s file: %v", pathToConfigJSON, err.Error())
	}
	// persistence specific configuration file
	config = new(types.UtilityConfig)
	err = json.Unmarshal(rawJSON[u.GetModuleName()], config)
	return
}

func (u *UtilityModule) InitGenesis(pathToGenesisJSON string) (genesis modules.IGenesis, err error) {
	return // No-op
}

func (u *UtilityModule) Start() error {
	return nil
}

func (u *UtilityModule) Stop() error {
	return nil
}

func (u *UtilityModule) GetModuleName() string {
	return UtilityModuleName
}

func (u *UtilityModule) SetBus(bus modules.Bus) {
	u.bus = bus
}

func (u *UtilityModule) GetBus() modules.Bus {
	if u.bus == nil {
		log.Fatalf("Bus is not initialized")
	}
	return u.bus
}
