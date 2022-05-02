package genesis

import (
	"encoding/hex"
	"fmt"

	"log"
	"sync"

	"github.com/pokt-network/pocket/shared/config"
)

// TODO(team): This structure is a proxy into the current / active state of the network
// containing information such as the current validator map. As a next step, we can move
// all of this data over into the persistence module.
type NodeState struct {
	GenesisState *GenesisState

	BlockHeight      uint64
	AppHash          string                // TODO: Why not call this a BlockHash or StateHash? SHould it be a []byte or string?
	ValidatorMap     map[string]*Validator // TODO: Need to update this on every validator pause/stake/unstake/etc.
	TotalVotingPower uint64                // TODO: Need to update this on every send transaction.
}

// The pocket state singleton.
var state *NodeState

// Used to load the state when the singleton is created.
var once sync.Once

// Used to update the state. All exported functions should lock this when they are called and defer an unlock.
var lock = &sync.Mutex{}

// TODO(team): Passing both config and genesis to `GetNodeState` is a hack and only used for integration purposes
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

func (ps *NodeState) loadStateFromConfig(cfg *config.Config) {
	lock.Lock()
	defer lock.Unlock()

	if cfg.Persistence != nil && len(cfg.Persistence.DataDir) > 0 {
		panic("[TODO] Load p2p state from persistence not supported. Only supporting loading p2p state from genesis file for now.")
	} else if len(cfg.Genesis) > 0 {
		log.Println("Loading state from Genesis")
		ps.loadStateFromGenesis(cfg)
		return
	}

	log.Fatalf("[TODO] Config must not be nil when initializing the pocket state. ...")
}

func (ps *NodeState) loadStateFromGenesis(cfg *config.Config) {
	genesis, err := PocketGenesisFromFileOrJSON(cfg.Genesis)
	if err != nil {
		log.Fatalf("Failed to load genesis: %v", err)
	}

	if genesis.GenesisStateConfig != nil {
		log.Println("Loading state from `genesis_state_configs`")
		genesisState, _, _, _, _, err := NewGenesisState(genesis.GenesisStateConfig)
		if err != nil {
			log.Fatalf("Failed to generate genesis: %v", err)
		}

		if genesis.Validators != nil {
			genesisState.Validators = GetValidators(genesis.Validators)
		}

		*ps = NodeState{
			GenesisState: genesisState,
			BlockHeight:  0,
			AppHash:      genesis.AppHash,
			ValidatorMap: ValidatorListToMap(genesisState.Validators),
		}
	} else {
		log.Println("Loading state from json file data")
		*ps = NodeState{
			GenesisState: nil,
			BlockHeight:  0,
			AppHash:      genesis.AppHash,
			ValidatorMap: ValidatorListToMap(GetValidators(genesis.Validators)),
		}
	}

	ps.recomputeTotalVotingPower()
}

func ValidatorListToMap(validators []*Validator) (m map[string]*Validator) {
	m = make(map[string]*Validator, len(validators))
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

func (ps *NodeState) PrintGlobalState() {
	fmt.Printf("\tGLOBAL STATE: (BlockHeight, PrevAppHash, # Validators, TotalVotingPower) is: (%d, %s, %d, %d)\n", ps.BlockHeight, ps.AppHash, len(ps.ValidatorMap), ps.TotalVotingPower)
}

func (ps *NodeState) UpdateAppHash(appHash string) {
	ps.AppHash = appHash
}

func (ps *NodeState) UpdateBlockHeight(blockHeight uint64) {
	ps.BlockHeight = blockHeight
}
