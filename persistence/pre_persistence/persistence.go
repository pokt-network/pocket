package pre_persistence

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"

	"github.com/jordanorelli/lexnum"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"
)

const (
	FirstSavePointKeyName             = "first_savepoint_key"
	DeletedPrefixKeyName              = "first_savepoint_key"
	BlockPrefixName                   = "first_savepoint_key"
	TransactionKeyPrefixName          = "transaction/"
	PoolPrefixKeyName                 = "pool/"
	AccountPrefixKeyName              = "account/"
	AppPrefixKeyName                  = "app/"
	UnstakingAppPrefixKeyName         = "unstaking_app/"
	ServiceNodePrefixKeyName          = "service_node/"
	UnstakingServiceNodePrefixKeyName = "unstaking_service_node/"
	FishermanPrefixKeyName            = "fisherman/"
	UnstakingFishermanPrefixKeyName   = "unstaking_fisherman/"
	ValidatorPrefixKeyName            = "validator/"
	UnstakingValidatorPrefixKeyName   = "unstaking_validator/"
	ParamsPrefixKeyName               = "params/"
)

var (
	FirstSavePointKey                                        = []byte(FirstSavePointKeyName)
	DeletedPrefixKey                                         = []byte(DeletedPrefixKeyName)
	BlockPrefix                                              = []byte(BlockPrefixName)
	TransactionKeyPrefix                                     = []byte(TransactionKeyPrefixName)
	PoolPrefixKey                                            = []byte(PoolPrefixKeyName)
	AccountPrefixKey                                         = []byte(AccountPrefixKeyName)
	AppPrefixKey                                             = []byte(AppPrefixKeyName)
	UnstakingAppPrefixKey                                    = []byte(UnstakingAppPrefixKeyName)
	ServiceNodePrefixKey                                     = []byte(ServiceNodePrefixKeyName)
	UnstakingServiceNodePrefixKey                            = []byte(UnstakingServiceNodePrefixKeyName)
	FishermanPrefixKey                                       = []byte(FishermanPrefixKeyName)
	UnstakingFishermanPrefixKey                              = []byte(UnstakingFishermanPrefixKeyName)
	ValidatorPrefixKey                                       = []byte(ValidatorPrefixKeyName)
	UnstakingValidatorPrefixKey                              = []byte(UnstakingValidatorPrefixKeyName)
	ParamsPrefixKey                                          = []byte(ParamsPrefixKeyName)
	_                             modules.PersistenceModule  = &PrePersistenceModule{}
	_                             modules.PersistenceContext = &PrePersistenceContext{}
	elenEncoder                                              = lexnum.NewEncoder('=', '-')
)

type PrePersistenceModule struct { // TODO make private if possible
	bus modules.Bus

	CommitDB *memdb.DB
	Mempool  types.Mempool
	Cfg      *config.Config
}

func NewPrePersistenceModule(commitDB *memdb.DB, mempool types.Mempool, cfg *config.Config) *PrePersistenceModule {
	return &PrePersistenceModule{CommitDB: commitDB, Mempool: mempool, Cfg: cfg}
}

func (m *PrePersistenceModule) NewContext(height int64) (modules.PersistenceContext, error) {
	newDB := NewMemDB()
	it := m.CommitDB.NewIterator(&util.Range{
		Start: HeightKey(height, nil),
		Limit: HeightKey(height+1, nil),
	})
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		err := newDB.Put(KeyFromHeightKey(it.Key()), it.Value())
		if err != nil {
			return nil, err
		}
	}
	context := &PrePersistenceContext{
		Height: height,
		Parent: m,
		DBs:    make([]*memdb.DB, 0),
	}
	context.DBs = append(context.DBs, newDB)
	return context, nil
}

func (m *PrePersistenceModule) GetCommitDB() *memdb.DB {
	return m.CommitDB
}

type PrePersistenceContext struct {
	Height     int64
	Parent     modules.PersistenceModule
	SavePoints map[string]int // TODO save points not entirely implemented. Happy path only for now, rollbacks for later
	DBs        []*memdb.DB
}

func (m *PrePersistenceContext) GetLatestBlockHeight() (uint64, error) {
	return uint64(m.Height), nil
}

// ExportState Unused but high potential for usefulness for telemetry
func (m *PrePersistenceContext) ExportState() (*GenesisState, types.Error) {
	var err error
	state := &GenesisState{}
	state.Validators, err = m.GetAllValidators(m.Height)
	if err != nil {
		return nil, types.ErrGetAllValidators(err)
	}
	state.Apps, err = m.GetAllApps(m.Height)
	if err != nil {
		return nil, types.ErrGetAllApps(err)
	}
	state.Fishermen, err = m.GetAllFishermen(m.Height)
	if err != nil {
		return nil, types.ErrGetAllFishermen(err)
	}
	state.ServiceNodes, err = m.GetAllServiceNodes(m.Height)
	if err != nil {
		return nil, types.ErrGetAllServiceNodes(err)
	}
	state.Pools, err = m.GetAllPools(m.Height)
	if err != nil {
		return nil, types.ErrGetAllPools(err)
	}
	state.Accounts, err = m.GetAllAccounts(m.Height)
	if err != nil {
		return nil, types.ErrGetAllAccounts(err)
	}
	state.Params, err = m.GetParams(m.Height)
	if err != nil {
		return nil, types.ErrGetAllParams(err)
	}
	return state, nil
}

// NewSavePoint Create a save point
// Needed for atomic rollbacks in the case of failed transactions during proposal or blocks during validation
func (m *PrePersistenceContext) NewSavePoint(bytes []byte) error {
	index := len(m.DBs)
	newDB := NewMemDB()
	if index == 0 {
		return fmt.Errorf("%s", "zero length mock persistence context")
	}
	src := m.DBs[index-1]
	if err := CopyMemDB(src, newDB); err != nil {
		return err
	}
	m.SavePoints[hex.EncodeToString(bytes)] = index
	m.DBs = append(m.DBs, newDB)
	return nil
}

// RollbackToSavePoint Rollback save point
// Needed in the case of failed transactions during proposal or blocks during validation
func (m *PrePersistenceContext) RollbackToSavePoint(bytes []byte) error {
	rollbackIndex, ok := m.SavePoints[hex.EncodeToString(bytes)]
	if !ok {
		return fmt.Errorf("save point not found")
	}
	toDelete := make([]string, 0)
	// rollback savepoints map
	for key, i := range m.SavePoints {
		if i > rollbackIndex {
			toDelete = append(toDelete, key)
		}
	}
	for _, key := range toDelete {
		delete(m.SavePoints, key)
	}
	// rollback
	m.DBs = m.DBs[:rollbackIndex]
	return nil
}

// AppHash creates a unique hash based on the global state object
// NOTE: AppHash is an inefficient, arbitrary, mock implementation that enables the functionality
// TODO written for replacement, taking any and all better implementation suggestions - even if a temporary measure
// Assigned Andrewnguyen22 / Iajz
func (m *PrePersistenceContext) AppHash() ([]byte, error) {
	result := make([]byte, 0)
	index := len(m.DBs) - 1
	db := m.DBs[index]
	it := db.NewIterator(&util.Range{})
	for valid := it.First(); valid; valid = it.Next() {
		result = append(result, it.Value()...)
		// chunk into 100000 byte segments
		if len(result) >= 100000 {
			result = crypto.SHA3Hash(result)
		}
	}
	it.Release()
	// potential for double hash here
	return crypto.SHA3Hash(result), nil
}

// Reset to the first save point
func (m *PrePersistenceContext) Reset() error {
	return m.RollbackToSavePoint(FirstSavePointKey)
}

// Commit the KV pairs to the parent (commit) db
func (m *PrePersistenceContext) Commit() error {
	index := len(m.DBs) - 1
	db := m.DBs[index]
	it := db.NewIterator(&util.Range{})
	for valid := it.First(); valid; valid = it.Next() {
		if err := m.Parent.GetCommitDB().Put(HeightKey(m.Height, it.Key()), it.Value()); err != nil {
			return err
		}
	}
	it.Release()
	m.Release()
	parentIt := m.Parent.GetCommitDB().NewIterator(&util.Range{
		Start: HeightKey(m.Height, nil),
		Limit: PrefixEndBytes(HeightKey(m.Height, nil)),
	})
	parentIt.First()
	m.Height = m.Height + 1
	// copy over the entire last height
	for ; parentIt.Valid(); parentIt.Next() {
		newKey := HeightKey(m.Height, KeyFromHeightKey(parentIt.Key()))
		if err := m.Parent.GetCommitDB().Put(newKey, parentIt.Value()); err != nil {
			return err
		}
	}
	parentIt.Release()
	return nil
}

func (m *PrePersistenceContext) Release() {
	m.SavePoints = nil
	for _, db := range m.DBs {
		db.Reset()
	}
	m.DBs = nil
	return
}

// Store returns the latest 'app state' db object
func (m *PrePersistenceContext) Store() *memdb.DB {
	i := len(m.DBs) - 1
	if i < 0 {
		panic(fmt.Errorf("zero length mock persistence context"))
	}
	return m.DBs[i]
}

func (m *PrePersistenceContext) GetHeight() (int64, error) {
	return m.Height, nil
}

func (m *PrePersistenceContext) GetBlockHash(height int64) ([]byte, error) {
	db := m.Store()
	block := Block{}
	key := append(BlockPrefix, Int64ToBytes(height)...)
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(val, &block); err != nil {
		return nil, err
	}
	return []byte(block.BlockHeader.Hash), nil
}

func (m *PrePersistenceContext) TransactionExists(transactionHash string) bool {
	db := m.Store()
	return db.Contains(append(TransactionKeyPrefix, []byte(transactionHash)...))
}

func NewMemDB() *memdb.DB {
	return memdb.New(comparer.DefaultComparer, 100000)
}

func CopyMemDB(src, dest *memdb.DB) error {
	it := src.NewIterator(&util.Range{})
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		err := dest.Put(it.Key(), it.Value())
		if err != nil {
			return err
		}
	}
	return nil
}

func HeightKey(height int64, k []byte) (key []byte) {
	keyString := fmt.Sprintf("%s/%s", elenEncoder.EncodeInt(int(height)), k)
	return []byte(keyString)
}

func KeyFromHeightKey(heightKey []byte) (key []byte) {
	k := strings.SplitN(string(heightKey), "/", 2)[1]
	return []byte(k)
}

// PrefixEndBytes : Returns the 'END RANGE' or LIMIT for a prefix; Commonly used in KV range functions
func PrefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for {
		if end[len(end)-1] != byte(255) {
			end[len(end)-1]++
			break
		} else {
			end = end[:len(end)-1]
			if len(end) == 0 {
				end = nil
				break
			}
		}
	}
	return end
}
