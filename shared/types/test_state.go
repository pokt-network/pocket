package types

import (
	"log"
	"sync"

	"github.com/pokt-network/pocket/shared/config"
)

// TODO(hack): This is a singleton that can be used as a view into the genesis
// file - since state is not currently being loaded from disk. This structure will
// either be removed or redesigned altogether.
type TestState struct {
	BlockHeight      uint64 // The current block height of the chain (updated through load, state sync, normal operation, etc)
	AppHash          string // TODO(discuss): Why not call this a BlockHash or StateHash? Should it be a []byte or string?
	ValidatorMap     ValMap // TODO(olshansky): Need to update this on every validator operation(pause, stake, unstake, etc)
	TotalVotingPower uint64 // TODO(design): Need to update this on every send transaction.
}

// The pocket state singleton.
var state *TestState

// Used to load the state when the singleton is created.
var once sync.Once

// Used to update the state. All exported functions should lock this when they are called and defer an unlock.
var lock = &sync.Mutex{}

// TODO(hack): This is a singleton that requires a config to be passed in the first time it
// is created, and subsequent calls should simply pass in a nil to get the current state.
func GetTestState(cfg *config.Config) *TestState {
	once.Do(func() {
		if state == nil && cfg == nil {
			log.Fatalf("TestState has not been initialized yet, so a config must be specified.")
		}

		if state != nil && cfg != nil {
			log.Fatalf("TestState has already been initialized, so a config should not be specified.")
		}

		state = &TestState{}
		state.loadStateFromConfig(cfg)
	})

	return state
}

func (ps *TestState) loadStateFromConfig(cfg *config.Config) {
	lock.Lock()
	defer lock.Unlock()

	persistenceConfig := cfg.Persistence
	if persistenceConfig == nil || len(persistenceConfig.DataDir) == 0 {
		// TODO(design): Load p2p state from persistence. Only supporting loading p2p state from genesis file for now.
		ps.loadStateFromGenesis(cfg)
		return
	}

	log.Fatalf("[TODO] Config must not be nil when initializing the pocket state. ...")
}

func (ps *TestState) loadStateFromGenesis(cfg *config.Config) {
	genesis, err := PocketGenesisFromFileOrJSON(cfg.Genesis)
	if err != nil {
		log.Fatalf("Failed to load genesis: %v", err)
	}

	if len(cfg.PrivateKey) == 0 {
		log.Fatalf("[TODO] Private key must be set when initializing the pocket state. ...")
	}

	*ps = TestState{
		BlockHeight:      0,
		AppHash:          genesis.AppHash,
		ValidatorMap:     ValidatorListToMap(genesis.Validators),
		TotalVotingPower: 0, // Value is compute below in `recomputeTotalVotingPower`
	}

	ps.recomputeTotalVotingPower()
}

func (ps *TestState) recomputeTotalVotingPower() {
	ps.TotalVotingPower = 0
	for _, v := range ps.ValidatorMap {
		ps.TotalVotingPower += v.UPokt
	}
}

func (ps *TestState) PrintGlobalState() {
	log.Printf("\tGLOBAL STATE: (BlockHeight, PrevAppHash, # Validators, TotalVotingPower) is: (%d, %s, %d, %d)\n", ps.BlockHeight, ps.AppHash, len(ps.ValidatorMap), ps.TotalVotingPower)
}

func (ps *TestState) UpdateAppHash(appHash string) {
	ps.AppHash = appHash
}

func (ps *TestState) UpdateBlockHeight(blockHeight uint64) {
	ps.BlockHeight = blockHeight
}
