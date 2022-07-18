package genesis

// TODO(team): Consolidate this with `shared/genesis.go`

import (
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
)

func GenesisStateFromFile(file string) (*GenesisState, error) {
	jsonBlob, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("couldn't read genesis file: %w", err)
	}
	genesis, err := GenesisStateFromJson(jsonBlob)
	if err != nil {
		return nil, fmt.Errorf("error generating genesis state from json: %w", err)
	}
	return genesis, nil
}

func GenesisStateFromJson(jsonBlob []byte) (*GenesisState, error) {
	genesis := GenesisState{}
	if err := json.Unmarshal(jsonBlob, &genesis); err != nil {
		return nil, err
	}
	if err := genesis.Validate(); err != nil {
		return nil, err
	}
	return &genesis, nil
}

// TODO: Validate each field in GenesisState
func (genesis *GenesisState) Validate() error {
	return nil
}

// See the explanation here for the need of this function: https://stackoverflow.com/a/73015992/768439
func (source *GenesisSource) UnmarshalJSON(data []byte) error {
	protojson.Unmarshal(data, source)
	return nil
}
