package nodestate

import (
	"log"
	"testing"

	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func TestLoadStateFromGenesis(t *testing.T) {
	pk, err := cryptoPocket.GeneratePrivateKey()
	require.NoError(t, err)
	cfg := &config.Config{
		GenesisSource: &genesis.GenesisSource{
			Source: &genesis.GenesisSource_State{
				State: genesisState(),
			},
		},
		PrivateKey: pk.(cryptoPocket.Ed25519PrivateKey),
	}
	ResetNodeState(t)
	state := GetNodeState(cfg)
	require.Equal(t, 4, len(state.ValidatorMap))
}

func TestLoadStateFrompersistence(t *testing.T) {
	log.Println("[TODO] Load state from the persistence module")
}

func genesisState() *genesis.GenesisState {
	return &genesis.GenesisState{
		Validators: []*genesis.Validator{
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
			{
				Address:         []byte("e3c1b362c0df36f6b370b8b1479b67dad96392b2"),
				PublicKey:       []byte("6b79c57e6a095239282c04818e96112f3f03a4001ba97a564c23852a3f1ea5fc"),
				Paused:          false,
				Status:          2,
				ServiceUrl:      "",
				StakedTokens:    "",
				MissedBlocks:    0,
				PausedHeight:    0,
				UnstakingHeight: 0,
				Output:          []byte("e3c1b362c0df36f6b370b8b1479b67dad96392b2"),
			},
			{
				Address:         []byte("fa4d86c3b551aa6cd7c3759d040c037ef2c6379f"),
				PublicKey:       []byte("cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc"),
				Paused:          false,
				Status:          2,
				ServiceUrl:      "",
				StakedTokens:    "",
				MissedBlocks:    0,
				PausedHeight:    0,
				UnstakingHeight: 0,
				Output:          []byte("fa4d86c3b551aa6cd7c3759d040c037ef2c6379f"),
			},
		},
	}
}
