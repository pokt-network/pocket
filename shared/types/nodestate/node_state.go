package nodestate

import (
	"encoding/hex"
	"os"
	"testing"

	"log"
	"sync"

	"github.com/matryer/resync"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

// TODO(team): This structure is a proxy into the current / active state of the network
// containing information such as the current validator map. As a next step, we can move
// all of this data over into the persistence module.
type NodeState struct {
	GenesisState *genesis.GenesisState

	ValidatorMap     map[string]*genesis.Validator // TODO: Need to update this on every validator pause/stake/unstake/etc.
	TotalVotingPower uint64                        // TODO: Need to update this on every send transaction.
}

// The pocket state singleton.
var state *NodeState

// Used to load the state when the singleton is created.
var once resync.Once

// Used to update the state. All exported functions should lock this when they are called and defer an unlock.
var lock = &sync.Mutex{}

// REFACTOR(pocket/issues/84): look into refactoring this to use a DI framework and delete the use
// of Singleton's altogether.
func GetNodeState(cfg *config.Config) *NodeState {
	once.Do(func() {
		if state == nil && cfg == nil {
			log.Fatalf("NodeState has not been initialized yet, so a config must be specified.")
		}

		if state != nil && cfg != nil {
			log.Fatalf("NodeState has already been initialized, so a config should not be specified.")
		}

		state = &NodeState{}
		state.loadStateFromConfig(cfg)
	})

	return state
}

// HACK(olshansky): The NodeState singleton is being complex but is outside the scope of current changes.
// For testing purposes, it needs to be reset because it exists in a global scope but multiple nodes
// are configured in parallel.
func ResetNodeState(_ *testing.T) {
	lock.Lock()
	defer lock.Unlock()
	once.Reset()
	state = nil
}

// TODO: Load state from persistent storage
func (ps *NodeState) loadStateFromConfig(cfg *config.Config) {
	lock.Lock()
	defer lock.Unlock()

	if cfg.GenesisSource != nil {
		log.Println("Loading state from Genesis")
		ps.loadStateFromGenesisSource(cfg.GenesisSource)
		return
	}

	log.Fatalf("Unsupported case to load state from config...")
}

func (ps *NodeState) loadStateFromGenesisSource(genesisSource *genesis.GenesisSource) {
	var genesisState *genesis.GenesisState
	var err error

	switch genesisSource.Source.(type) {
	case *genesis.GenesisSource_Config:
		genesisConfig := genesisSource.GetConfig()
		if genesisState, _, _, _, _, err = genesis.NewGenesisState(genesisConfig); err != nil {
			log.Fatalf("Failed to generate genesis state from configuration: %v", err)
		}
	case *genesis.GenesisSource_File:
		genesisFilePath := genesisSource.GetFile().Path
		if _, err := os.Stat(genesisFilePath); err != nil {
			log.Fatalf("Genesis file specified but not found %s", genesisFilePath)
		}
		if genesisState, err = genesis.GenesisStateFromFile(genesisFilePath); err != nil {
			log.Fatalf("Failed to load genesis state from file: %v", err)
		}
	case *genesis.GenesisSource_State:
		genesisState = genesisSource.GetState()
	default:
		log.Fatalf("Unsupported genesis source type: %v", genesisSource.Source)
	}

	*ps = NodeState{
		GenesisState: genesisState,
		ValidatorMap: ValidatorListToMap(genesisState.Validators),
	}
	ps.recomputeTotalVotingPower()
}

func ValidatorListToMap(validators []*genesis.Validator) (m map[string]*genesis.Validator) {
	m = make(map[string]*genesis.Validator, len(validators))
	for _, v := range validators {
		m[hex.EncodeToString(v.Address)] = v
	}
	return
}

// TODO(olshansky): Re-implement this when properly implementing leader election
func (ps *NodeState) recomputeTotalVotingPower() {
	ps.TotalVotingPower = 0
	// for _, v := range ps.ValidatorMap {
	// 	ps.TotalVotingPower += v.UPokt
	// }
}
