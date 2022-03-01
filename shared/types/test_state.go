package types

import (
	"log"
	"sync"

	"github.com/pokt-network/pocket/shared/config"
	pcrypto "github.com/pokt-network/pocket/shared/crypto"
)

// TODO(hack by olshansky): This is a singleton that can be used as a view into the genesis file (since state is not currently being loaded)
// from disk. It will be removed in the future but currently enables continued work.
type TestState struct {
	BlockHeight      uint64
	AppHash          string // TODO(discuss): Why not call this a BlockHash or StateHash? Should it be a []byte or string?
	ValidatorMap     ValMap // TODO(olshansky): Need to update this on every validator pause/stake/unstake/etc.
	TotalVotingPower uint64 // TODO(team): Need to update this on every send transaction.

	PrivateKey pcrypto.PrivateKey

	Config config.Config // TODO(hack): Should we store this here?
}

// The pocket state singleton.
var state *TestState

// Used to load the state when the singleton is created.
var once sync.Once

// Used to update the state. All exported functions should lock this when they are called and defer an unlock.
var lock = &sync.Mutex{}

func GetTestState() *TestState {
	once.Do(func() {
		state = &TestState{}
	})

	return state
}

func (ps *TestState) LoadStateFromConfig(cfg *config.Config) {
	lock.Lock()
	defer lock.Unlock()

	persistenceConfig := cfg.Persistence
	if persistenceConfig == nil || len(persistenceConfig.DataDir) == 0 {
		log.Println("[TODO] Load p2p state from persistence. Only supporting loading p2p state from genesis file for now.")
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
		BlockHeight:  0,
		ValidatorMap: ValidatorListToMap(genesis.Validators),

		PrivateKey: cfg.PrivateKey,

		Config: *cfg,
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
