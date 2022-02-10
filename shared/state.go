package shared

import (
	"fmt"
	"log"
	"sync"

	"pocket/consensus/pkg/config"
	"pocket/consensus/pkg/types"
)

// TODO: Return this as a singleton!
// Consider using the singleton pattern? https://medium.com/golang-issue/how-singleton-pattern-works-with-golang-2fdd61cd5a7f

type pocketState struct {
	BlockHeight      uint64
	AppHash          string       // TODO: Why not call this a BlockHash or StateHash? SHould it be a []byte or string?
	ValidatorMap     types.ValMap // TODO: Need to update this on every validator pause/stake/unstake/etc.
	TotalVotingPower uint64       // TODO: Need to update this on every send transaction.

	PublicKey types.PublicKey
	Address   string

	Config          config.Config // TODO: Should we store this here?
	ConsensusParams config.ConsensusParams

	// Sync State?
	// Node Type?
}

// The pocket state singleton.
var state *pocketState

// Used to load the state when the singleton is created.
var once sync.Once

// Used to update the state. All exported functions should lock this when they are called and defer an unlock.
var lock = &sync.Mutex{}

func GetPocketState() *pocketState {
	once.Do(func() {
		state = &pocketState{}
	})

	return state
}

func (ps *pocketState) LoadStateFromConfig(cfg *config.Config) {
	lock.Lock()
	defer lock.Unlock()

	persistenceConfig := cfg.Persistence
	if persistenceConfig == nil || len(persistenceConfig.DataDir) == 0 {
		log.Println("[TODO] Load pocket state from persistence. Only supporting loading pocket state from genesis file for now.")
		ps.loadStateFromGenesis(cfg)
		return
	}

	log.Fatalf("[TODO] Config must not be nil when initializing the pocket state. ...")
}

func (ps *pocketState) AddValidator(v *types.Validator) {
	lock.Lock()
	defer lock.Unlock()

	ps.ValidatorMap[v.NodeId] = v
	ps.TotalVotingPower += v.UPokt
}

func (ps *pocketState) RemoveValidator(nodeId types.NodeId) {
	lock.Lock()
	defer lock.Unlock()

	v, ok := ps.ValidatorMap[nodeId]
	if !ok {
		log.Println("[WARN] Trying to remove a validator not found in the PocketState: ", nodeId)
		return
	}
	ps.TotalVotingPower -= v.UPokt
	delete(ps.ValidatorMap, nodeId)
}

func (ps *pocketState) loadStateFromGenesis(cfg *config.Config) {
	genesis, err := config.PocketGenesisFromFileOrJSON(cfg.Genesis)
	if err != nil {
		log.Fatalf("Failed to load genesis: %v", err)
	}

	if cfg.PrivateKey == nil {
		log.Fatalf("[TODO] Private key must be set when initializing the pocket state. ...")
	}

	*ps = pocketState{
		BlockHeight:  0,
		AppHash:      genesis.AppHash,
		ValidatorMap: types.ValidatorListToMap(genesis.Validators),

		PublicKey: cfg.PrivateKey.Public(),
		Address:   types.AddressFromKey(cfg.PrivateKey.Public()),

		Config:          *cfg,
		ConsensusParams: *genesis.ConsensusParams,
	}

	ps.recomputeTotalVotingPower()
}

func (ps *pocketState) recomputeTotalVotingPower() {
	ps.TotalVotingPower = 0
	for _, v := range ps.ValidatorMap {
		ps.TotalVotingPower += v.UPokt
	}
}

func (ps *pocketState) PrintGlobalState() {
	fmt.Printf("\tGLOBAL STATE: (BlockHeight, PrevAppHash, # Validators, TotalVotingPower) is: (%d, %s, %d, %d)\n", ps.BlockHeight, ps.AppHash, len(ps.ValidatorMap), ps.TotalVotingPower)
}

func (ps *pocketState) UpdateAppHash(appHash string) {
	ps.AppHash = appHash
}

func (ps *pocketState) UpdateBlockHeight(blockHeight uint64) {
	ps.BlockHeight = blockHeight
}
