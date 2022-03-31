package pre_persistence

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/crypto"
)

// TODO: Return this as a singleton!
// Consider using the singleton pattern? https://medium.com/golang-issue/how-singleton-pattern-works-with-golang-2fdd61cd5a7f

type TestState struct {
	BlockHeight      uint64
	AppHash          string // TODO: Why not call this a BlockHash or StateHash? SHould it be a []byte or string?
	ValidatorMap     ValMap // TODO: Need to update this on every validator pause/stake/unstake/etc.
	TotalVotingPower uint64 // TODO: Need to update this on every send transaction.

	PublicKey crypto.PublicKey
	Address   string

	Config config.Config // TODO: Should we store this here?
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

	*ps = TestState{
		BlockHeight:  0,
		ValidatorMap: ValidatorListToMap(genesis.Validators),

		PublicKey: cfg.PrivateKey.PublicKey(),
		Address:   cfg.PrivateKey.Address().String(),

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
	fmt.Printf("\tGLOBAL STATE: (BlockHeight, PrevAppHash, # Validators, TotalVotingPower) is: (%d, %s, %d, %d)\n", ps.BlockHeight, ps.AppHash, len(ps.ValidatorMap), ps.TotalVotingPower)
}

func (ps *TestState) UpdateAppHash(appHash string) {
	ps.AppHash = appHash
}

func (ps *TestState) UpdateBlockHeight(blockHeight uint64) {
	ps.BlockHeight = blockHeight
}

// A stable monotimally increasing integer used for
// identification, consensus, distributed key generation.
// Please note that this is needed and cannot be substituted
// for by address or public keys.
type NodeId uint32

type ValMap map[NodeId]*TestVal

type TestVal struct {
	NodeId     NodeId   `json:"node_id"`
	Address    string   `json:"address"`
	PublicKey  string   `json:"public_key"`
	PrivateKey string   `json:"private_Key"`
	Jailed     bool     `json:"jailed"` // TODO: Integrate with utility to update this.
	UPokt      uint64   `json:"upokt"`  // TODO: Integrate with utility to update this.
	Host       string   `json:"host"`
	Port       uint32   `json:"port"`
	DebugPort  uint32   `json:"debug_port"`
	Chains     []string `json:"chains"` // TODO: Integrate with utility to update this.

}

func (v *TestVal) Validate() error {
	// log.Println("[TODO] Validator config validation is not implemented yet.")
	return nil
}

func ValidatorListToMap(validators []*TestVal) (m ValMap) {
	m = make(ValMap, len(validators))
	for _, v := range validators {
		m[v.NodeId] = v
	}
	return
}

type TestGenesis struct {
	GenesisTime time.Time  `json:"genesis_time"`
	AppHash     string     `json:"app_hash"`
	Validators  []*TestVal `json:"validators"`
}

// TODO: This is a temporary hack that can load Genesis from a single string
// that may be either a JSON blob or a file.
func PocketGenesisFromFileOrJSON(fileOrJson string) (*TestGenesis, error) {
	if _, err := os.Stat(fileOrJson); err == nil {
		return PocketGenesisFromFile(fileOrJson)
	}
	return PocketGenesisFromJSON([]byte(fileOrJson))
}

func PocketGenesisFromFile(file string) (*TestGenesis, error) {
	jsonBlob, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("couldn't read TestGenesis file: %w", err)
	}
	genesis, err := PocketGenesisFromJSON(jsonBlob)
	if err != nil {
		return nil, fmt.Errorf("error reading TestGenesis at %s: %w", file, err)
	}
	return genesis, nil
}

func PocketGenesisFromJSON(jsonBlob []byte) (*TestGenesis, error) {
	genesis := TestGenesis{}
	if err := json.Unmarshal(jsonBlob, &genesis); err != nil {
		return nil, err
	}

	if err := genesis.Validate(); err != nil {
		return nil, err
	}

	return &genesis, nil
}

func (genesis *TestGenesis) Validate() error {
	if genesis.GenesisTime.IsZero() {
		return fmt.Errorf("GenesisTime cannot be zero")
	}
	// TODO: validate each account.
	if len(genesis.Validators) == 0 {
		return fmt.Errorf("genesis must contain at least one validator")
	}
	for _, validator := range genesis.Validators {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("validator in genesis is invalid: %w", err)
		}
	}

	return nil
}
