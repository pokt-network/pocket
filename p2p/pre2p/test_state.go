package pre2p

import (
	"log"
	"sync"

	"github.com/pokt-network/pocket/p2p/pre2p/types"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/crypto"
)

// TODO(discuss): This whole structure can potentially be removed altogether in the future once mainline has a functioning end-to-end implementation.
type TestState struct {
	BlockHeight      uint64
	AppHash          string       // TODO(discuss): Why not call this a BlockHash or StateHash? Should it be a []byte or string?
	ValidatorMap     types.ValMap // TODO(olshansky): Need to update this on every validator pause/stake/unstake/etc.
	TotalVotingPower uint64       // TODO(team): Need to update this on every send transaction.

	PublicKey crypto.PublicKey
	Address   string

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

	if cfg.PrivateKey == "" {
		log.Fatalf("[TODO] Private key must be set when initializing the pocket state. ...")
	}
	pk, err := crypto.NewPrivateKey(cfg.PrivateKey)
	if err != nil {
		panic(err)
	}
	*ps = TestState{
		BlockHeight:  0,
		ValidatorMap: types.ValidatorListToMap(genesis.Validators),

		PublicKey: pk.PublicKey(),
		Address:   pk.Address().String(),

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
