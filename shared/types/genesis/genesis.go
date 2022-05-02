package genesis

// TODO(team): Consolidate this with `shared/genesis.go`

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Genesis struct {
	// TODO(olshansky): Discuss this structure with Andrew.
	GenesisStateConfig *NewGenesisStateConfigs `json:"genesis_state_configs"`

	GenesisTime time.Time                         `json:"genesis_time"`
	AppHash     string                            `json:"app_hash"`
	Validators  []*ValidatorJsonCompatibleWrapper `json:"validators"`
}

// TODO: This is a temporary hack that can load Genesis from a single string
// that may be either a JSON blob or a file.
func PocketGenesisFromFileOrJSON(fileOrJson string) (*Genesis, error) {
	if _, err := os.Stat(fileOrJson); err == nil {
		return PocketGenesisFromFile(fileOrJson)
	}
	return PocketGenesisFromJSON([]byte(fileOrJson))
}

func PocketGenesisFromFile(file string) (*Genesis, error) {
	jsonBlob, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("couldn't read Genesis file: %w", err)
	}
	genesis, err := PocketGenesisFromJSON(jsonBlob)
	if err != nil {
		return nil, fmt.Errorf("error reading Genesis at %s: %w", file, err)
	}
	return genesis, nil
}

func PocketGenesisFromJSON(jsonBlob []byte) (*Genesis, error) {
	genesis := Genesis{}
	if err := json.Unmarshal(jsonBlob, &genesis); err != nil {
		return nil, err
	}
	if err := genesis.Validate(); err != nil {
		return nil, err
	}
	return &genesis, nil
}

func (genesis *Genesis) Validate() error {
	if genesis.GenesisTime.IsZero() {
		return fmt.Errorf("GenesisTime cannot be zero")
	}

	// TODO: validate each account.
	if len(genesis.Validators) == 0 && (genesis.GenesisStateConfig == nil || genesis.GenesisStateConfig.NumValidators == 0) {
		return fmt.Errorf("genesis must contain at least one validator")
	}

	if len(genesis.AppHash) == 0 {
		return fmt.Errorf("Genesis app hash cannot be zero")
	}

	for _, validator := range genesis.Validators {
		if err := validator.ValidateBasic(); err != nil {
			return fmt.Errorf("validator in genesis is invalid: %w", err)
		}
	}

	return nil
}
