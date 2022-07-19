package genesis

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisStateFromJson(t *testing.T) {
	genesisJson := `{
		"validators": [
		  {
			"address": "71f8be163036c0da94f188bb817d77691869ccff5932059f3c398f2fb92fa08b",
			"public_key": "b1f804dabc68274c1233995c5a9119b56935bcdd83b7de07ec726dcedc4e9ce7",
			"paused": false,
			"status": 0,
			"service_url": "validator.com",
			"staked_tokens": "42",
			"missed_blocks": 0,
			"paused_height": 0,
			"unstaking_height": 0,
			"output": "71f8be163036c0da94f188bb817d77691869ccff5932059f3c398f2fb92fa08b"
		  }
		],
		"accounts": [],
		"pools": [],
		"fisherman": [],
		"service_nodes": [],
		"apps": [],
		"params": {}
	  }`

	genesisState, err := GenesisStateFromJson([]byte(genesisJson))
	require.NoError(t, err)

	require.Equal(t, len(genesisState.Validators), 1)
	require.Equal(t, len(genesisState.Accounts), 0)
	require.Equal(t, len(genesisState.Pools), 0)
	require.Equal(t, len(genesisState.Fishermen), 0)
	require.Equal(t, len(genesisState.ServiceNodes), 0)
	require.Equal(t, len(genesisState.Apps), 0)
}

func TestGenesisStateFromConfigSource(t *testing.T) {
	genesisSource := &GenesisSource{
		Source: &GenesisSource_Config{
			Config: &GenesisConfig{
				NumValidators:   4,
				NumApplications: 0,
				NumFisherman:    0,
				NumServicers:    0,
			},
		},
	}
	genesisState, err := GenesisStateFromGenesisSource(genesisSource)
	require.NoError(t, err)

	require.Equal(t, len(genesisState.Validators), 4)
	require.Equal(t, len(genesisState.Accounts), 5) // The 4 validators above + 1 DAO account
	require.Equal(t, len(genesisState.Pools), 6)    // There are 6 hard coded pools
	require.Equal(t, len(genesisState.Fishermen), 0)
	require.Equal(t, len(genesisState.ServiceNodes), 0)
	require.Equal(t, len(genesisState.Apps), 0)
}

func TestGenesisStateFromFileSource(t *testing.T) {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	genesisFileSource := &GenesisSource{
		Source: &GenesisSource_File{
			File: &GenesisFile{
				Path: filepath.Join(basepath, "./test_artifacts/test_genesis.json"),
			},
		},
	}

	genesisState, err := GenesisStateFromGenesisSource(genesisFileSource)
	require.NoError(t, err)

	require.Equal(t, len(genesisState.Validators), 4)
	require.Equal(t, len(genesisState.Accounts), 0)
	require.Equal(t, len(genesisState.Pools), 0)
	require.Equal(t, len(genesisState.Fishermen), 0)
	require.Equal(t, len(genesisState.ServiceNodes), 0)
	require.Equal(t, len(genesisState.Apps), 0)
}

func TestGenesisStateFromStateSource(t *testing.T) {
	genesisStateSource := &GenesisSource{
		Source: &GenesisSource_State{
			State: &GenesisState{
				Validators: []*Validator{
					{
						Address:         []byte("6f1e5b61ed9a821457aa6b4d7c2a2b37715ffb16"),
						PublicKey:       []byte("9be3287795907809407e14439ff198d5bfc7dce6f9bc743cb369146f610b4801"),
						Paused:          false,
						Status:          2, // TODO: Add an enum of constants or something else to make this clear.
						ServiceUrl:      "",
						StakedTokens:    "",
						MissedBlocks:    0,
						PausedHeight:    0,
						UnstakingHeight: 0,
						Output:          []byte("6f1e5b61ed9a821457aa6b4d7c2a2b37715ffb16"),
					},
					{
						Address:         []byte("db0743e2dcba9ebf2419bde0881beea966689a26"),
						PublicKey:       []byte("dadbd184a2d526f1ebdd5c06fdad9359b228759b4d7f79d66689fa254aad8546"),
						Paused:          false,
						Status:          2,
						ServiceUrl:      "",
						StakedTokens:    "",
						MissedBlocks:    0,
						PausedHeight:    0,
						UnstakingHeight: 0,
						Output:          []byte("db0743e2dcba9ebf2419bde0881beea966689a26"),
					},
				},
				Accounts:     []*Account{},
				Pools:        []*Pool{},
				Fishermen:    []*Fisherman{},
				ServiceNodes: []*ServiceNode{},
				Apps:         []*App{},
				Params:       &Params{},
			},
		},
	}

	genesisState, err := GenesisStateFromGenesisSource(genesisStateSource)
	require.NoError(t, err)

	require.Equal(t, len(genesisState.Validators), 2)
	require.Equal(t, len(genesisState.Accounts), 0)
	require.Equal(t, len(genesisState.Pools), 0)
	require.Equal(t, len(genesisState.Fishermen), 0)
	require.Equal(t, len(genesisState.ServiceNodes), 0)
	require.Equal(t, len(genesisState.Apps), 0)

}
