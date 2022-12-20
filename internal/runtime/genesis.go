package runtime

import (
	"encoding/json"
	"os"

	consTypes "github.com/pokt-network/pocket/internal/consensus/types"
	persTypes "github.com/pokt-network/pocket/internal/persistence/types"
	"github.com/pokt-network/pocket/internal/shared/modules"
)

var _ modules.GenesisState = &runtimeGenesis{}

type runtimeGenesis struct {
	ConsensusGenesisState   *consTypes.ConsensusGenesisState   `json:"consensus_genesis_state"`
	PersistenceGenesisState *persTypes.PersistenceGenesisState `json:"persistence_genesis_state"`
}

func NewGenesis(
	consensusGenesisState modules.ConsensusGenesisState,
	persistenceGenesisState modules.PersistenceGenesisState,
) *runtimeGenesis {
	return &runtimeGenesis{
		ConsensusGenesisState:   consensusGenesisState.(*consTypes.ConsensusGenesisState),
		PersistenceGenesisState: persistenceGenesisState.(*persTypes.PersistenceGenesisState),
	}
}

func (g *runtimeGenesis) GetPersistenceGenesisState() modules.PersistenceGenesisState {
	return g.PersistenceGenesisState
}

func (g *runtimeGenesis) GetConsensusGenesisState() modules.ConsensusGenesisState {
	return g.ConsensusGenesisState
}

func parseGenesisJSON(genesisPath string) (g *runtimeGenesis, err error) {
	data, err := os.ReadFile(genesisPath)
	if err != nil {
		return
	}

	// general genesis file
	g = new(runtimeGenesis)
	err = json.Unmarshal(data, &g)
	return
}
