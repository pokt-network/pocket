package genesis

// TODO(team): Consolidate this with `shared/genesis.go`

import (
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
)

// Specialized accounts. See the utility specification for more details.
const (
	ServiceNodeStakePoolName = "SERVICE_NODE_STAKE_POOL"
	AppStakePoolName         = "APP_STAKE_POOL"
	ValidatorStakePoolName   = "VALIDATOR_STAKE_POOL"
	FishermanStakePoolName   = "FISHERMAN_STAKE_POOL"
	DAOPoolName              = "DAO_POOL"
	FeePoolName              = "FEE_POOL"
)

func GenesisStateFromGenesisSource(genesisSource *GenesisSource) (genesisState *GenesisState, err error) {
	switch genesisSource.Source.(type) {
	case *GenesisSource_Config:
		genesisConfig := genesisSource.GetConfig()
		if genesisState, _, _, _, _, err = NewGenesisState(genesisConfig); err != nil {
			return nil, fmt.Errorf("failed to generate genesis state from configuration: %v", err)
		}
	case *GenesisSource_File:
		genesisFilePath := genesisSource.GetFile().Path
		if _, err := os.Stat(genesisFilePath); err != nil {
			return nil, fmt.Errorf("genesis file specified but not found %s", genesisFilePath)
		}
		if genesisState, err = GenesisStateFromFile(genesisFilePath); err != nil {
			return nil, fmt.Errorf("failed to load genesis state from file: %v", err)
		}
	case *GenesisSource_State:
		genesisState = genesisSource.GetState()
	default:
		return nil, fmt.Errorf("unsupported genesis source type: %v", genesisSource.Source)
	}

	return
}

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
