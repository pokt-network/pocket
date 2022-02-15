package types

import (
	"crypto"
	"fmt"
	"github.com/mindstand/gogm/v2"
	"log"
	"pocket/shared/config"
	crypto2 "pocket/shared/crypto"
	"sync"
)

// TODO: Return this as a singleton!
// Consider using the singleton pattern? https://medium.com/golang-issue/how-singleton-pattern-works-with-golang-2fdd61cd5a7f

type ConsensusState struct {
	BlockHeight      uint64
	AppHash          string // TODO: Why not call this a BlockHash or StateHash? SHould it be a []byte or string?
	ValidatorMap     ValMap // TODO: Need to update this on every validator pause/stake/unstake/etc.
	TotalVotingPower uint64 // TODO: Need to update this on every send transaction.

	PublicKey crypto.PublicKey
	Address   string

	Config          config.Config // TODO: Should we store this here?
	ConsensusParams ConsensusParams

	// Sync State?
	// Node Type?
}

type ConsensusNodeState struct {
	NodeId   NodeId `gogm:"name=NodeId"`
	Height   uint64 `gogm:"name=Height"` // TODO: Change to proper type
	Round    uint8  `gogm:"name=Round"`  // TODO: Change to proper type
	Step     uint8  `gogm:"name=Step"`   // TODO: Change to proper type
	IsLeader bool   `gogm:"name=IsLeader"`
	LeaderId NodeId `gogm:"name=Leader"`

	gogm.BaseNode // Provides required node fields for neo4j DB
}

// The pocket state singleton.
var state *ConsensusState

// Used to load the state when the singleton is created.
var once sync.Once

// Used to update the state. All exported functions should lock this when they are called and defer an unlock.
var lock = &sync.Mutex{}

func GetPocketState() *ConsensusState {
	once.Do(func() {
		state = &ConsensusState{}
	})

	return state
}

func (ps *ConsensusState) LoadStateFromConfig(cfg *config.Config) {
	lock.Lock()
	defer lock.Unlock()

	persistenceConfig := cfg.Persistence
	if persistenceConfig == nil || len(persistenceConfig.DataDir) == 0 {
		log.Println("[TODO] Load consensus state from persistence. Only supporting loading consensus state from genesis file for now.")
		ps.loadStateFromGenesis(cfg)
		return
	}

	log.Fatalf("[TODO] Config must not be nil when initializing the pocket state. ...")
}

func (ps *ConsensusState) loadStateFromGenesis(cfg *config.Config) {
	genesis, err := PocketGenesisFromFileOrJSON(cfg.Genesis)
	if err != nil {
		log.Fatalf("Failed to load genesis: %v", err)
	}

	if cfg.PrivateKey == "" {
		log.Fatalf("[TODO] Private key must be set when initializing the pocket state. ...")
	}

	pk, err := crypto2.NewPrivateKey(cfg.PrivateKey)
	if err != nil {
		panic(err)
	}
	*ps = ConsensusState{
		BlockHeight:  0,
		AppHash:      genesis.AppHash,
		ValidatorMap: ValidatorListToMap(genesis.Validators),

		PublicKey: pk.PublicKey(),
		Address:   pk.Address().String(),

		Config:          *cfg,
		ConsensusParams: *genesis.ConsensusParams,
	}

	ps.recomputeTotalVotingPower()
}

func (ps *ConsensusState) recomputeTotalVotingPower() {
	ps.TotalVotingPower = 0
	for _, v := range ps.ValidatorMap {
		ps.TotalVotingPower += v.UPokt
	}
}

func (ps *ConsensusState) PrintGlobalState() {
	fmt.Printf("\tGLOBAL STATE: (BlockHeight, PrevAppHash, # Validators, TotalVotingPower) is: (%d, %s, %d, %d)\n", ps.BlockHeight, ps.AppHash, len(ps.ValidatorMap), ps.TotalVotingPower)
}

func (ps *ConsensusState) UpdateAppHash(appHash string) {
	ps.AppHash = appHash
}

func (ps *ConsensusState) UpdateBlockHeight(blockHeight uint64) {
	ps.BlockHeight = blockHeight
}
