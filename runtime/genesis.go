package runtime

import (
	"encoding/json"
	"os"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	typesPers "github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.GenesisState = &runtimeGenesis{}

type runtimeGenesis struct {
	ConsensusGenesisState   *typesCons.ConsensusGenesisState   `json:"consensus_genesis_state"`
	PersistenceGenesisState *typesPers.PersistenceGenesisState `json:"persistence_genesis_state"`
}

func NewGenesis(
	consensusGenesisState modules.ConsensusGenesisState,
	persistenceGenesisState modules.PersistenceGenesisState,
) *runtimeGenesis {
	return &runtimeGenesis{
		ConsensusGenesisState:   consensusGenesisState.(*typesCons.ConsensusGenesisState),
		PersistenceGenesisState: persistenceGenesisState.(*typesPers.PersistenceGenesisState),
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
