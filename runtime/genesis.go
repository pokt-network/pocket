package runtime

import (
	"encoding/json"
	"os"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	typesPers "github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ modules.ConsensusGenesisState = &Genesis{}
var _ modules.PersistenceGenesisState = &Genesis{}

type Genesis struct {
	ConsensusGenesisState   *typesCons.ConsensusGenesisState   `json:"consensus_genesis_state"`
	PersistenceGenesisState *typesPers.PersistenceGenesisState `json:"persistence_genesis_state"`
}

func (g *Genesis) ToShared() modules.GenesisState {
	return modules.GenesisState{
		PersistenceGenesisState: g,
		ConsensusGenesisState:   g,
	}
}

func ParseGenesisJSON(genesisPath string) (genesis *Genesis, err error) {
	data, err := os.ReadFile(genesisPath)
	if err != nil {
		return
	}

	// general genesis file
	genesis = new(Genesis)
	err = json.Unmarshal(data, &genesis)
	return
}

// modules.ConsensusGenesisState

func (g *Genesis) GetGenesisTime() *timestamppb.Timestamp {
	return g.ConsensusGenesisState.GenesisTime
}
func (g *Genesis) GetChainId() string {
	return g.ConsensusGenesisState.ChainId
}
func (g *Genesis) GetMaxBlockBytes() uint64 {
	return g.ConsensusGenesisState.MaxBlockBytes
}

// modules.PersistenceGenesisState

func (g *Genesis) GetAccs() []modules.Account {
	return g.PersistenceGenesisState.GetAccs()
}

func (g *Genesis) GetAccPools() []modules.Account {
	return g.PersistenceGenesisState.GetAccPools()
}

func (g *Genesis) GetApps() []modules.Actor {
	return g.PersistenceGenesisState.GetApps()
}

func (g *Genesis) GetVals() []modules.Actor {
	return g.PersistenceGenesisState.GetVals()
}

func (g *Genesis) GetFish() []modules.Actor {
	return g.PersistenceGenesisState.GetFish()
}

func (g *Genesis) GetNodes() []modules.Actor {
	return g.PersistenceGenesisState.GetNodes()
}

func (g *Genesis) GetParameters() modules.Params {
	return g.PersistenceGenesisState.GetParameters()
}
