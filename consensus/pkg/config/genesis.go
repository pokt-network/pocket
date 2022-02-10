// See https://github.com/pokt-network/pocket-network-genesis as a reference
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"pocket/consensus/pkg/types"
)

type PocketGenesis struct {
	GenesisTime time.Time          `json:"genesis_time"`
	AppHash     string             `json:"app_hash"`
	Accounts    []*types.Accounts  `json:"accounts"`
	Validators  []*types.Validator `json:"validators"`

	ConsensusParams *ConsensusParams `json:"consensus_params"`
}

// TODO: This is a temporary hack that can load Genesis from a single string
// that may be either a JSON blob or a file.
func PocketGenesisFromFileOrJSON(fileOrJson string) (*PocketGenesis, error) {
	if _, err := os.Stat(fileOrJson); err == nil {
		return PocketGenesisFromFile(fileOrJson)
	}
	return PocketGenesisFromJSON([]byte(fileOrJson))
}

func PocketGenesisFromFile(file string) (*PocketGenesis, error) {
	jsonBlob, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("couldn't read PocketGenesis file: %w", err)
	}
	genesis, err := PocketGenesisFromJSON(jsonBlob)
	if err != nil {
		return nil, fmt.Errorf("error reading PocketGenesis at %s: %w", file, err)
	}
	return genesis, nil
}

func PocketGenesisFromJSON(jsonBlob []byte) (*PocketGenesis, error) {
	genesis := PocketGenesis{}
	if err := json.Unmarshal(jsonBlob, &genesis); err != nil {
		return nil, err
	}

	if err := genesis.Validate(); err != nil {
		return nil, err
	}

	return &genesis, nil
}

func (genesis *PocketGenesis) Validate() error {
	if genesis.GenesisTime.IsZero() {
		return fmt.Errorf("GenesisTime cannot be zero")
	}

	if err := genesis.ConsensusParams.Validate(); err != nil {
		return fmt.Errorf("ConsensusParams genesis error: %w", err)
	}

	if len(genesis.Accounts) == 0 {
		return fmt.Errorf("genesis must contain at least one account")
	}
	for _, account := range genesis.Accounts {
		if err := account.Validate(); err != nil {
			return fmt.Errorf("account in genesis is invalid: %w", err)
		}
	}

	// TODO: validate each account.
	if len(genesis.Validators) == 0 {
		return fmt.Errorf("genesis must contain at least one validator")
	}
	for _, validator := range genesis.Validators {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("validator in genesis is invalid: %w", err)
		}
	}

	return nil
}
