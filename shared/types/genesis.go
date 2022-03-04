package types

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Come back to this.
// TODO(olshansky): This is an interim genesis structure that will be replaced with a real one. It is the bare minimum for prototyping purposes.
type Genesis struct {
	GenesisTime time.Time    `json:"genesis_time"`
	AppHash     string       `json:"app_hash"`
	Validators  []*Validator `json:"validators"`
}

// TODO(olshansky): Temporary hack that can load Genesis from a single string
// that may be either a JSON blob or a file. Should be removed in the future.
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

	if len(genesis.Validators) == 0 {
		return fmt.Errorf("Genesis must contain at least one validator")
	}

	if len(genesis.AppHash) == 0 {
		return fmt.Errorf("Genesis app hash cannot be zero")
	}

	for _, validator := range genesis.Validators {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("validator in genesis is invalid: %w", err)
		}
	}

	return nil
}
