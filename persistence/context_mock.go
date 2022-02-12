package persistence

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"pocket/shared/modules"
	"pocket/shared/typespb"
	"pocket/utility/shared/crypto"
	"pocket/utility/utility/test"
	"pocket/utility/utility/types"
	"strings"

	"github.com/jordanorelli/lexnum"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/memdb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"
)

func NewMockMempool() types.Mempool {
	return types.NewMempool(1000000, 1000000)
}

var (
	defaultChains           = []string{"0001"}
	defaultChainsEdited     = []string{"0002"}
	defaultServiceURL       = "https://foo.bar"
	defaultServiceURLEdited = "https://bar.foo"
	defaultStakeBig         = big.NewInt(1000000000000000)
	defaultStake            = types.BigIntToString(defaultStakeBig)
	defaultAccountbalance   = defaultStake
	defaultStakeStatus      = int32(2)
)

func NewMockGenesisState(numOfValidators, numOfApplications, numOfFisherman, numOfServiceNodes int) (state *typespb.GenesisState, validatorKeys, appKeys, serviceNodeKeys, fishKeys []crypto.PrivateKey, err error) {
	state = &typespb.GenesisState{}
	validatorKeys = make([]crypto.PrivateKey, numOfValidators)
	appKeys = make([]crypto.PrivateKey, numOfApplications)
	fishKeys = make([]crypto.PrivateKey, numOfFisherman)
	serviceNodeKeys = make([]crypto.PrivateKey, numOfServiceNodes)
	for i := range validatorKeys {
		pk, _ := crypto.GeneratePrivateKey()
		v := &typespb.Validator{
			Status:       2,
			ServiceURL:   defaultServiceURL,
			StakedTokens: defaultStake,
		}
		v.Address = pk.Address()
		v.PublicKey = pk.PublicKey().Bytes()
		v.Output = v.Address
		state.Validators = append(state.Validators, v)
		state.Accounts = append(state.Accounts, &typespb.Account{
			Address: v.Address,
			Amount:  defaultAccountbalance,
		})
		validatorKeys[i] = pk
	}
	for i := range appKeys {
		pk, _ := crypto.GeneratePrivateKey()
		app := &typespb.App{
			Status:       defaultStakeStatus,
			Chains:       defaultChains,
			StakedTokens: defaultStake,
		}
		app.Address = pk.Address()
		app.PublicKey = pk.PublicKey().Bytes()
		app.Output = app.Address
		state.Apps = append(state.Apps, app)
		state.Accounts = append(state.Accounts, &typespb.Account{
			Address: app.Address,
			Amount:  defaultAccountbalance,
		})
		appKeys[i] = pk
	}
	for i := range serviceNodeKeys {
		pk, _ := crypto.GeneratePrivateKey()
		sn := &typespb.ServiceNode{
			Status:       defaultStakeStatus,
			ServiceURL:   defaultServiceURL,
			Chains:       defaultChains,
			StakedTokens: defaultStake,
		}
		sn.Address = pk.Address()
		sn.PublicKey = pk.PublicKey().Bytes()
		sn.Output = sn.Address
		state.ServiceNodes = append(state.ServiceNodes, sn)
		state.Accounts = append(state.Accounts, &typespb.Account{
			Address: sn.Address,
			Amount:  defaultAccountbalance,
		})
		serviceNodeKeys[i] = pk
	}
	for i := range fishKeys {
		pk, _ := crypto.GeneratePrivateKey()
		fish := &typespb.Fisherman{
			Status:       defaultStakeStatus,
			Chains:       defaultChains,
			ServiceURL:   defaultServiceURL,
			StakedTokens: defaultStake,
		}
		fish.Address = pk.Address()
		fish.PublicKey = pk.PublicKey().Bytes()
		fish.Output = fish.Address
		state.Fishermen = append(state.Fishermen, fish)
		state.Accounts = append(state.Accounts, &typespb.Account{
			Address: fish.Address,
			Amount:  defaultAccountbalance,
		})
		fishKeys[i] = pk
	}
	// state.Params = test.DefaultParams()
	dao, err := test.NewPool(types.DAOPoolName, &test.Account{
		Address: test.DefaultDAOPool.Address(),
		Amount:  types.BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	valStakePool, err := test.NewPool(types.ValidatorStakePoolName, &test.Account{
		Address: test.DefaultValidatorStakePool.Address(),
		Amount:  types.BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	appStakePool, err := test.NewPool(types.AppStakePoolName, &test.Account{
		Address: test.DefaultAppStakePool.Address(),
		Amount:  types.BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	fishStakePool, err := test.NewPool(types.FishermanStakePoolName, &test.Account{
		Address: test.DefaultFishermanStakePool.Address(),
		Amount:  types.BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	serNodeStakePool, err := test.NewPool(types.ServiceNodeStakePoolName, &test.Account{
		Address: test.DefaultServiceNodeStakePool.Address(),
		Amount:  types.BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	fee, err := test.NewPool(types.FeePoolName, &test.Account{
		Address: test.DefaultFeeCollector.Address(),
		Amount:  types.BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	pOwnerAddress := test.DefaultParamsOwner.Address()
	state.Accounts = append(state.Accounts, &typespb.Account{
		Address: pOwnerAddress,
		Amount:  defaultAccountbalance,
	})
	// state.Pools = append(state.Pools, dao)
	// state.Pools = append(state.Pools, fee)
	// state.Pools = append(state.Pools, serNodeStakePool)
	// state.Pools = append(state.Pools, fishStakePool)
	// state.Pools = append(state.Pools, appStakePool)
	// state.Pools = append(state.Pools, valStakePool)
	fmt.Println(dao, fee, serNodeStakePool, fishStakePool, appStakePool, valStakePool)
	return
}

var (
	firstSavePointKey             = []byte("first_savepoint_key")
	DeletedPrefixKey              = []byte("deleted/")
	BlockPrefix                   = []byte("block/")
	TransactionKeyPrefix          = []byte("transaction/")
	PoolPrefixKey                 = []byte("pool/")
	AccountPrefixKey              = []byte("account/")
	AppPrefixKey                  = []byte("app/")
	UnstakingAppPrefixKey         = []byte("unstaking_app/")
	ServiceNodePrefixKey          = []byte("service_node/")
	UnstakingServiceNodePrefixKey = []byte("unstaking_service_node/")
	FishermanPrefixKey            = []byte("fisherman/")
	UnstakingFishermanPrefixKey   = []byte("unstaking_fisherman/")
	ValidatorPrefixKey            = []byte("validator/")
	UnstakingValidatorPrefixKey   = []byte("unstaking_validator/")
	ParamsPrefixKey               = []byte("params/")
	//_                             bus.PersistenceModule  = &MockPersistenceModule{}
	//_                             bus.PersistenceContext = &MockPersistenceContext{}
)

//type MockPersistenceModule struct {
//	CommitDB *memdb.DB

//}

type MockPersistenceContext struct {
	modules.PersistenceContext

	Height     int64
	Parent     modules.PersistenceModule
	SavePoints map[string]int
	DBs        []*memdb.DB
}

func (m *MockPersistenceContext) ExportState() (*typespb.GenesisState, types.Error) {
	var err error
	state := &typespb.GenesisState{}
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

func (m *MockPersistenceContext) NewSavePoint(bytes []byte) error {
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

func (m *MockPersistenceContext) RollbackToSavePoint(bytes []byte) error {
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

func (m *MockPersistenceContext) AppHash() ([]byte, error) {
	// copy over the new values
	index := len(m.DBs) - 1
	db := m.DBs[index]
	it := db.NewIterator(&util.Range{})
	result := make([]byte, 0)
	for ; it.Valid(); it.Next() {
		result = append(result, it.Value()...)
		if len(result) >= 100000 {
			result = crypto.SHA3Hash(result)
		}
	}
	it.Release()
	return crypto.SHA3Hash(result), nil
}

func (m *MockPersistenceContext) Reset() error {
	return m.RollbackToSavePoint(firstSavePointKey)
}

func (m *MockPersistenceContext) Commit() error {
	parentIt := m.Parent.GetCommitDB().NewIterator(&util.Range{})
	// copy over the entire last height
	for ; parentIt.Valid(); parentIt.Next() {
		newKey := HeightKey(m.Height, KeyFromHeightKey(parentIt.Key()))
		if err := m.Parent.GetCommitDB().Put(newKey, parentIt.Value()); err != nil {
			return err
		}
	}
	parentIt.Release()
	// copy over the new values
	index := len(m.DBs) - 1
	db := m.DBs[index]
	it := db.NewIterator(&util.Range{})
	for ; it.Valid(); it.Next() {
		if err := m.Parent.GetCommitDB().Put(HeightKey(m.Height, it.Key()), it.Value()); err != nil {
			return err
		}
	}
	it.Release()
	m.Release()
	return nil
}

func (m *MockPersistenceContext) Release() {
	m.SavePoints = nil
	for _, db := range m.DBs {
		db.Reset()
	}
	m.DBs = nil
	return
}

func (m *MockPersistenceContext) Store() *memdb.DB {
	i := len(m.DBs) - 1
	if i < 0 {
		panic(fmt.Errorf("zero length mock persistence context"))
	}
	return m.DBs[i]
}

func (m *MockPersistenceContext) GetHeight() (int64, error) {
	return m.Height, nil
}

func (m *MockPersistenceContext) GetBlockHash(height int64) ([]byte, error) {
	db := m.Store()
	block := typespb.Block{}
	key := append(BlockPrefix, []byte(fmt.Sprintf("%d", height))...)
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(val, &block); err != nil {
		return nil, err
	}
	return []byte(block.BlockHeader.Hash), nil
}

func (m *MockPersistenceContext) TransactionExists(transactionHash string) bool {
	db := m.Store()
	return db.Contains(append(TransactionKeyPrefix, []byte(transactionHash)...))
}

func (m *MockPersistenceContext) AddPoolAmount(name string, amount string) error {
	cdc := types.UtilityCodec()
	p := typespb.Pool{}
	db := m.Store()
	key := append(PoolPrefixKey, []byte(name)...)
	val, err := db.Get(key)
	if err != nil {
		return err
	}
	err = cdc.Unmarshal(val, &p)
	if err != nil {
		return err
	}
	s, err := types.StringToBigInt(p.Account.Amount)
	if err != nil {
		return err
	}
	s2, err := types.StringToBigInt(amount)
	if err != nil {
		return err
	}
	s.Add(s, s2)
	p.Account.Amount = types.BigIntToString(s)
	bz, err := cdc.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) SubtractPoolAmount(name string, amount string) error {
	cdc := types.UtilityCodec()
	p := typespb.Pool{}
	db := m.Store()
	key := append(PoolPrefixKey, []byte(name)...)
	val, err := db.Get(key)
	if err != nil {
		return err
	}
	err = cdc.Unmarshal(val, &p)
	if err != nil {
		return err
	}
	s, err := types.StringToBigInt(p.Account.Amount)
	if err != nil {
		return err
	}
	s2, err := types.StringToBigInt(amount)
	if err != nil {
		return err
	}
	s.Sub(s, s2)
	p.Account.Amount = types.BigIntToString(s)
	bz, err := cdc.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) InsertPool(name string, address []byte, amount string) error {
	cdc := types.UtilityCodec()
	p := typespb.Pool{
		Name: name,
		Account: &typespb.Account{
			Address: address,
			Amount:  amount,
		},
	}
	db := m.Store()
	key := append(PoolPrefixKey, []byte(name)...)
	bz, err := cdc.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) SetPoolAmount(name string, amount string) error {
	cdc := types.UtilityCodec()
	p := typespb.Pool{}
	db := m.Store()
	key := append(PoolPrefixKey, []byte(name)...)
	val, err := db.Get(key)
	if err != nil {
		return err
	}
	err = cdc.Unmarshal(val, &p)
	if err != nil {
		return err
	}
	p.Account.Amount = amount
	bz, err := cdc.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) GetPoolAmount(name string) (amount string, err error) {
	cdc := types.UtilityCodec()
	p := typespb.Pool{}
	db := m.Store()
	key := append(PoolPrefixKey, []byte(name)...)
	val, err := db.Get(key)
	if err != nil {
		return types.EmptyString, err
	}
	err = cdc.Unmarshal(val, &p)
	if err != nil {
		return types.EmptyString, err
	}
	return p.Account.Amount, nil
}

func (m *MockPersistenceContext) GetAllAccounts(height int64) (accs []*typespb.Account, err error) {
	cdc := types.UtilityCodec()
	accs = make([]*typespb.Account, 0)
	var it iterator.Iterator
	if height == m.Height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: AccountPrefixKey,
			Limit: PrefixEndBytes(AccountPrefixKey),
		})
	} else {
		key := HeightKey(height, AccountPrefixKey)
		it = m.Parent.GetCommitDB().NewIterator(&util.Range{
			Start: key,
			Limit: PrefixEndBytes(key),
		})
	}
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		acc := typespb.Account{}
		if err := cdc.Unmarshal(bz, &acc); err != nil {
			return nil, err
		}
		accs = append(accs, &acc)
	}
	return
}

func (m *MockPersistenceContext) GetAllPools(height int64) (pools []*typespb.Pool, err error) {
	cdc := types.UtilityCodec()
	pools = make([]*typespb.Pool, 0)
	var it iterator.Iterator
	if height == m.Height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: PoolPrefixKey,
			Limit: PrefixEndBytes(PoolPrefixKey),
		})
	} else {
		key := HeightKey(height, PoolPrefixKey)
		it = m.Parent.GetCommitDB().NewIterator(&util.Range{
			Start: key,
			Limit: PrefixEndBytes(key),
		})
	}
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		p := typespb.Pool{}
		if err := cdc.Unmarshal(bz, &p); err != nil {
			return nil, err
		}
		pools = append(pools, &p)
	}
	return
}

func (m *MockPersistenceContext) AddAccountAmount(address []byte, amount string) error {
	cdc := types.UtilityCodec()
	account := &typespb.Account{}
	db := m.Store()
	key := append(AccountPrefixKey, address...)
	val, err := db.Get(key)
	if err != nil {
		return err
	}
	err = cdc.Unmarshal(val, account)
	if err != nil {
		return err
	}
	s, err := types.StringToBigInt(account.Amount)
	if err != nil {
		return err
	}
	s2, err := types.StringToBigInt(amount)
	if err != nil {
		return err
	}
	s.Add(s, s2)
	account.Amount = types.BigIntToString(s)
	bz, err := cdc.Marshal(account)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) SubtractAccountAmount(address []byte, amount string) error {
	cdc := types.UtilityCodec()
	account := &typespb.Account{}
	db := m.Store()
	key := append(AccountPrefixKey, address...)
	val, err := db.Get(key)
	if err != nil {
		return err
	}
	err = cdc.Unmarshal(val, account)
	if err != nil {
		return err
	}
	s, err := types.StringToBigInt(account.Amount)
	if err != nil {
		return err
	}
	s2, err := types.StringToBigInt(amount)
	if err != nil {
		return err
	}
	s.Sub(s, s2)
	account.Amount = types.BigIntToString(s)
	bz, err := cdc.Marshal(account)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) GetAccountAmount(address []byte) (string, error) {
	cdc := types.UtilityCodec()
	account := &typespb.Account{}
	db := m.Store()
	key := append(AccountPrefixKey, address...)
	val, err := db.Get(key)
	if err != nil {
		return types.EmptyString, err
	}
	err = cdc.Unmarshal(val, account)
	if err != nil {
		return types.EmptyString, err
	}
	return account.Amount, nil
}

func (m *MockPersistenceContext) SetAccount(address []byte, amount string) error {
	cdc := types.UtilityCodec()
	account := typespb.Account{
		Address: address,
		Amount:  amount,
	}
	db := m.Store()
	key := append(AccountPrefixKey, address...)
	bz, err := cdc.Marshal(&account)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) GetAppExists(address []byte) (exists bool, err error) {
	db := m.Store()
	key := append(AppPrefixKey, address...)
	if found := db.Contains(key); !found {
		return false, nil
	}
	bz, err := db.Get(key)
	if err != nil {
		return false, err
	}
	if bz == nil {
		return false, nil
	}
	if bytes.Contains(bz, DeletedPrefixKey) {
		return false, nil
	}
	return true, nil
}

func (m *MockPersistenceContext) GetApp(address []byte) (app *typespb.App, exists bool, err error) {
	app = &typespb.App{}
	db := m.Store()
	key := append(AppPrefixKey, address...)
	bz, err := db.Get(key)
	if err != nil {
		return nil, false, err
	}
	if bz == nil {
		return nil, false, nil
	}
	if bytes.Contains(bz, DeletedPrefixKey) {
		return nil, false, nil
	}
	if err = proto.Unmarshal(bz, app); err != nil {
		return nil, true, err
	}
	return app, true, nil
}

func (m *MockPersistenceContext) GetAllApps(height int64) (apps []*typespb.App, err error) {
	cdc := types.UtilityCodec()
	apps = make([]*typespb.App, 0)
	var it iterator.Iterator
	if height == m.Height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: AppPrefixKey,
			Limit: PrefixEndBytes(AppPrefixKey),
		})
	} else {
		key := HeightKey(height, AppPrefixKey)
		it = m.Parent.GetCommitDB().NewIterator(&util.Range{
			Start: key,
			Limit: PrefixEndBytes(key),
		})
	}
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		a := typespb.App{}
		if err := cdc.Unmarshal(bz, &a); err != nil {
			return nil, err
		}
		apps = append(apps, &a)
	}
	return
}

func (m *MockPersistenceContext) InsertApplication(address []byte, publicKey []byte, output []byte, paused bool, status int, maxRelays string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	if _, exists, _ := m.GetApp(address); exists {
		return fmt.Errorf("already exists in world state")
	}
	cdc := types.UtilityCodec()
	db := m.Store()
	key := append(AppPrefixKey, address...)
	app := typespb.App{
		Address:         address,
		PublicKey:       publicKey,
		Paused:          paused,
		Status:          int32(status),
		Chains:          chains,
		MaxRelays:       maxRelays,
		StakedTokens:    stakedTokens,
		PausedHeight:    uint64(pausedHeight),
		UnstakingHeight: unstakingHeight,
		Output:          output,
	}
	bz, err := cdc.Marshal(&app)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) UpdateApplication(address []byte, maxRelaysToAdd string, amountToAdd string, chainsToUpdate []string) error {
	app, exists, _ := m.GetApp(address)
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	cdc := types.UtilityCodec()
	db := m.Store()
	key := append(AppPrefixKey, address...)
	// compute new values
	stakedTokens, err := types.StringToBigInt(app.StakedTokens)
	if err != nil {
		return err
	}
	stakedTokensToAddI, err := types.StringToBigInt(amountToAdd)
	if err != nil {
		return err
	}
	stakedTokens.Add(stakedTokens, stakedTokensToAddI)
	maxRelays, err := types.StringToBigInt(app.MaxRelays)
	if err != nil {
		return err
	}
	maxRelaysToAddI, err := types.StringToBigInt(maxRelaysToAdd)
	if err != nil {
		return err
	}
	maxRelays.Add(maxRelays, maxRelaysToAddI)
	// update values
	app.MaxRelays = types.BigIntToString(maxRelays)
	app.StakedTokens = types.BigIntToString(stakedTokens)
	app.Chains = chainsToUpdate
	// marshal
	bz, err := cdc.Marshal(app)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) DeleteApplication(address []byte) error {
	if exists, _ := m.GetAppExists(address); !exists {
		return fmt.Errorf("does not exist in world state")
	}
	db := m.Store()
	key := append(AppPrefixKey, address...)
	return db.Put(key, DeletedPrefixKey)
}

func (m *MockPersistenceContext) GetAppsReadyToUnstake(height int64, status int) (apps []modules.UnstakingActor, err error) { // TODO delete unstaking
	db := m.Store()
	unstakingKey := append(UnstakingAppPrefixKey, []byte(fmt.Sprintf("%d", height))...)
	if has := db.Contains(unstakingKey); !has {
		return nil, nil
	}
	val, err := db.Get(unstakingKey)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return make([]modules.UnstakingActor, 0), nil
	}
	unstakingApps := typespb.UnstakingActors{}
	if err := proto.Unmarshal(val, &unstakingApps); err != nil {
		return nil, err
	}
	for _, app := range unstakingApps.UnstakingActors {
		apps = append(apps, app)
	}
	return
}

func (m *MockPersistenceContext) GetAppStatus(address []byte) (status int, err error) {
	app, exists, err := m.GetApp(address)
	if err != nil {
		return types.ZeroInt, err
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(app.Status), nil
}

func (m *MockPersistenceContext) SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	app, exists, err := m.GetApp(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	cdc := types.UtilityCodec()
	unstakingApps := typespb.UnstakingActors{}
	db := m.Store()
	key := append(AppPrefixKey, address...)
	app.UnstakingHeight = unstakingHeight
	app.Status = int32(status)
	// marshal
	bz, err := cdc.Marshal(app)
	if err != nil {
		return err
	}
	if err := db.Put(key, bz); err != nil {
		return err
	}
	unstakingKey := append(UnstakingAppPrefixKey, []byte(fmt.Sprintf("%d", unstakingHeight))...)
	if found := db.Contains(unstakingKey); found {
		val, err := db.Get(unstakingKey)
		if err != nil {
			return err
		}
		if err := proto.Unmarshal(val, &unstakingApps); err != nil {
			return err
		}
	}
	unstakingApps.UnstakingActors = append(unstakingApps.UnstakingActors, &typespb.UnstakingActor{
		Address:       app.Address,
		StakeAmount:   app.StakedTokens,
		OutputAddress: app.Output,
	})
	unstakingBz, err := cdc.Marshal(&unstakingApps)
	if err != nil {
		return err
	}
	return db.Put(unstakingKey, unstakingBz)
}

func (m *MockPersistenceContext) GetAppPauseHeightIfExists(address []byte) (int64, error) {
	app, exists, err := m.GetApp(address)
	if err != nil {
		return types.ZeroInt, nil
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int64(app.PausedHeight), nil
}

func (m *MockPersistenceContext) SetAppsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	db := m.Store()
	cdc := types.UtilityCodec()
	it := db.NewIterator(&util.Range{
		Start: AppPrefixKey,
		Limit: PrefixEndBytes(AppPrefixKey),
	})
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		app := typespb.App{}
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		if err := cdc.Unmarshal(bz, &app); err != nil {
			return err
		}
		if app.PausedHeight < uint64(pausedBeforeHeight) {
			app.UnstakingHeight = unstakingHeight
			app.Status = int32(status)
			if err := m.SetAppUnstakingHeightAndStatus(app.Address, app.UnstakingHeight, status); err != nil {
				return err
			}
			bz, err := cdc.Marshal(&app)
			if err != nil {
				return err
			}
			if err := db.Put(it.Key(), bz); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MockPersistenceContext) SetAppPauseHeight(address []byte, height int64) error {
	cdc := types.UtilityCodec()
	db := m.Store()
	app, exists, err := m.GetApp(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	app.Paused = true
	app.PausedHeight = uint64(height)
	bz, err := cdc.Marshal(app)
	if err != nil {
		return err
	}
	return db.Put(append(AppPrefixKey, address...), bz)
}

func (m *MockPersistenceContext) GetAppOutputAddress(operator []byte) (output []byte, err error) {
	app, exists, err := m.GetApp(operator)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("does not exist in world state")
	}
	return app.Output, nil
}

func (m *MockPersistenceContext) GetServiceNodeExists(address []byte) (exists bool, err error) {
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	if found := db.Contains(key); !found {
		return false, nil
	}
	bz, err := db.Get(key)
	if err != nil {
		return false, err
	}
	if bz == nil {
		return false, nil
	}
	if bytes.Contains(bz, DeletedPrefixKey) {
		return false, nil
	}
	return true, nil
}

func (m *MockPersistenceContext) GetServiceNode(address []byte) (sn *typespb.ServiceNode, exists bool, err error) {
	sn = &typespb.ServiceNode{}
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	bz, err := db.Get(key)
	if err != nil {
		return nil, false, err
	}
	if bz == nil {
		return nil, false, nil
	}
	if bytes.Contains(bz, DeletedPrefixKey) {
		return nil, false, nil
	}
	if err = proto.Unmarshal(bz, sn); err != nil {
		return nil, true, err
	}
	return sn, true, nil
}

func (m *MockPersistenceContext) GetAllServiceNodes(height int64) (sns []*typespb.ServiceNode, err error) {
	cdc := types.UtilityCodec()
	sns = make([]*typespb.ServiceNode, 0)
	var it iterator.Iterator
	if height == m.Height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: ServiceNodePrefixKey,
			Limit: PrefixEndBytes(ServiceNodePrefixKey),
		})
	} else {
		key := HeightKey(height, ServiceNodePrefixKey)
		it = m.Parent.GetCommitDB().NewIterator(&util.Range{
			Start: key,
			Limit: PrefixEndBytes(key),
		})
	}
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		sn := typespb.ServiceNode{}
		if err := cdc.Unmarshal(bz, &sn); err != nil {
			return nil, err
		}
		sns = append(sns, &sn)
	}
	return
}

func (m *MockPersistenceContext) InsertServiceNode(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	if _, exists, _ := m.GetServiceNode(address); exists {
		return fmt.Errorf("already exists in world state")
	}
	cdc := types.UtilityCodec()
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	sn := typespb.ServiceNode{
		Address:         address,
		PublicKey:       publicKey,
		Paused:          paused,
		Status:          int32(status),
		Chains:          chains,
		ServiceURL:      serviceURL,
		StakedTokens:    stakedTokens,
		PausedHeight:    uint64(pausedHeight),
		UnstakingHeight: unstakingHeight,
		Output:          output,
	}
	bz, err := cdc.Marshal(&sn)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) UpdateServiceNode(address []byte, serviceURL string, amountToAdd string, chains []string) error {
	sn, exists, _ := m.GetServiceNode(address)
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	cdc := types.UtilityCodec()
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	// compute new values
	stakedTokens, err := types.StringToBigInt(sn.StakedTokens)
	if err != nil {
		return err
	}
	stakedTokensToAddI, err := types.StringToBigInt(amountToAdd)
	if err != nil {
		return err
	}
	stakedTokens.Add(stakedTokens, stakedTokensToAddI)
	// update values
	sn.ServiceURL = serviceURL
	sn.StakedTokens = types.BigIntToString(stakedTokens)
	sn.Chains = chains
	// marshal
	bz, err := cdc.Marshal(sn)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) DeleteServiceNode(address []byte) error {
	if exists, _ := m.GetServiceNodeExists(address); !exists {
		return fmt.Errorf("does not exist in world state")
	}
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	return db.Put(key, DeletedPrefixKey)
}

func (m *MockPersistenceContext) GetServiceNodesReadyToUnstake(height int64, status int) (ServiceNodes []modules.UnstakingActor, err error) {
	db := m.Store()
	unstakingKey := append(UnstakingServiceNodePrefixKey, []byte(fmt.Sprintf("%d", height))...)
	if has := db.Contains(unstakingKey); !has {
		return nil, nil
	}
	val, err := db.Get(unstakingKey)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return make([]modules.UnstakingActor, 0), nil
	}
	unstakingActors := typespb.UnstakingActors{}
	if err := proto.Unmarshal(val, &unstakingActors); err != nil {
		return nil, err
	}
	for _, sn := range unstakingActors.UnstakingActors {
		ServiceNodes = append(ServiceNodes, sn)
	}
	return
}

func (m *MockPersistenceContext) GetServiceNodeStatus(address []byte) (status int, err error) {
	sn, exists, err := m.GetServiceNode(address)
	if err != nil {
		return types.ZeroInt, err
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(sn.Status), nil
}

func (m *MockPersistenceContext) SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	sn, exists, err := m.GetServiceNode(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	cdc := types.UtilityCodec()
	unstakingActors := typespb.UnstakingActors{}
	db := m.Store()
	key := append(ServiceNodePrefixKey, address...)
	sn.UnstakingHeight = unstakingHeight
	sn.Status = int32(status)
	// marshal
	bz, err := cdc.Marshal(sn)
	if err != nil {
		return err
	}
	if err := db.Put(key, bz); err != nil {
		return err
	}
	unstakingKey := append(UnstakingServiceNodePrefixKey, []byte(fmt.Sprintf("%d", unstakingHeight))...)
	if found := db.Contains(unstakingKey); found {
		val, err := db.Get(unstakingKey)
		if err != nil {
			return err
		}
		if err := proto.Unmarshal(val, &unstakingActors); err != nil {
			return err
		}
	}
	unstakingActors.UnstakingActors = append(unstakingActors.UnstakingActors, &typespb.UnstakingActor{
		Address:       sn.Address,
		StakeAmount:   sn.StakedTokens,
		OutputAddress: sn.Output,
	})
	unstakingBz, err := cdc.Marshal(&unstakingActors)
	if err != nil {
		return err
	}
	return db.Put(unstakingKey, unstakingBz)
}

func (m *MockPersistenceContext) GetServiceNodePauseHeightIfExists(address []byte) (int64, error) {
	sn, exists, err := m.GetServiceNode(address)
	if err != nil {
		return types.ZeroInt, nil
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int64(sn.PausedHeight), nil
}

func (m *MockPersistenceContext) SetServiceNodesStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	db := m.Store()
	cdc := types.UtilityCodec()
	it := db.NewIterator(&util.Range{
		Start: ServiceNodePrefixKey,
		Limit: PrefixEndBytes(ServiceNodePrefixKey),
	})
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		sn := typespb.ServiceNode{}
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		if err := cdc.Unmarshal(bz, &sn); err != nil {
			return err
		}
		if sn.PausedHeight < uint64(pausedBeforeHeight) {
			sn.UnstakingHeight = unstakingHeight
			sn.Status = int32(status)
			if err := m.SetServiceNodeUnstakingHeightAndStatus(sn.Address, sn.UnstakingHeight, status); err != nil {
				return err
			}
			bz, err := cdc.Marshal(&sn)
			if err != nil {
				return err
			}
			if err := db.Put(it.Key(), bz); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MockPersistenceContext) SetServiceNodePauseHeight(address []byte, height int64) error {
	cdc := types.UtilityCodec()
	db := m.Store()
	sn, exists, err := m.GetServiceNode(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	sn.Paused = true
	sn.PausedHeight = uint64(height)
	bz, err := cdc.Marshal(sn)
	if err != nil {
		return err
	}
	return db.Put(append(ServiceNodePrefixKey, address...), bz)
}

func (m *MockPersistenceContext) InitParams() error {
	cdc := types.UtilityCodec()
	db := m.Store()
	p := test.DefaultParams()
	bz, err := cdc.Marshal(p)
	if err != nil {
		return err
	}
	return db.Put(ParamsPrefixKey, bz)
}

func (m *MockPersistenceContext) GetParams(height int64) (p *typespb.Params, err error) {
	p = &typespb.Params{}
	cdc := types.UtilityCodec()
	var paramsBz []byte
	if height == m.Height {
		db := m.Store()
		paramsBz, err = db.Get(ParamsPrefixKey)
		if err != nil {
			return nil, err
		}
	} else {
		paramsBz, err = m.Parent.GetCommitDB().Get(HeightKey(height, ParamsPrefixKey))
		if err != nil {
			return nil, nil
		}
	}
	if err := cdc.Unmarshal(paramsBz, p); err != nil {
		return nil, err
	}
	return
}

func (m *MockPersistenceContext) GetServiceNodesPerSessionAt(height int64) (int, error) {
	params, err := m.GetParams(height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ServiceNodesPerSession), nil
}

func (m *MockPersistenceContext) GetServiceNodeCount(chain string, height int64) (int, error) {
	cdc := types.UtilityCodec()
	var it iterator.Iterator
	count := 0
	if m.Height == height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: ServiceNodePrefixKey,
			Limit: PrefixEndBytes(ServiceNodePrefixKey),
		})
	} else {
		it = m.Parent.GetCommitDB().NewIterator(&util.Range{
			Start: HeightKey(height, ServiceNodePrefixKey),
			Limit: HeightKey(height, PrefixEndBytes(ServiceNodePrefixKey)),
		})
	}
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		node := typespb.ServiceNode{}
		if err := cdc.Unmarshal(bz, &node); err != nil {
			return types.ZeroInt, err
		}
		for _, c := range node.Chains {
			if c == chain {
				count++
				break
			}
		}
	}
	return count, nil
}

func (m *MockPersistenceContext) GetServiceNodeOutputAddress(operator []byte) (output []byte, err error) {
	sn, exists, err := m.GetServiceNode(operator)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("does not exist in world state")
	}
	return sn.Output, nil
}

func (m *MockPersistenceContext) GetFishermanExists(address []byte) (exists bool, err error) {
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
	if found := db.Contains(key); !found {
		return false, nil
	}
	bz, err := db.Get(key)
	if err != nil {
		return false, err
	}
	if bz == nil {
		return false, nil
	}
	if bytes.Contains(bz, DeletedPrefixKey) {
		return false, nil
	}
	return true, nil
}

func (m *MockPersistenceContext) GetFisherman(address []byte) (fish *typespb.Fisherman, exists bool, err error) {
	fish = &typespb.Fisherman{}
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
	bz, err := db.Get(key)
	if err != nil {
		return nil, false, err
	}
	if bz == nil {
		return nil, false, nil
	}
	if bytes.Contains(bz, DeletedPrefixKey) {
		return nil, false, nil
	}
	if err = proto.Unmarshal(bz, fish); err != nil {
		return nil, true, err
	}
	return fish, true, nil
}

func (m *MockPersistenceContext) GetAllFishermen(height int64) (fishermen []*typespb.Fisherman, err error) {
	cdc := types.UtilityCodec()
	fishermen = make([]*typespb.Fisherman, 0)
	var it iterator.Iterator
	if height == m.Height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: FishermanPrefixKey,
			Limit: PrefixEndBytes(FishermanPrefixKey),
		})
	} else {
		key := HeightKey(height, FishermanPrefixKey)
		it = m.Parent.GetCommitDB().NewIterator(&util.Range{
			Start: key,
			Limit: PrefixEndBytes(key),
		})
	}
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		fish := typespb.Fisherman{}
		if err := cdc.Unmarshal(bz, &fish); err != nil {
			return nil, err
		}
		fishermen = append(fishermen, &fish)
	}
	return
}

func (m *MockPersistenceContext) InsertFisherman(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error {
	if _, exists, _ := m.GetFisherman(address); exists {
		return fmt.Errorf("already exists in world state")
	}
	cdc := types.UtilityCodec()
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
	fish := typespb.Fisherman{
		Address:         address,
		PublicKey:       publicKey,
		Paused:          paused,
		Status:          int32(status),
		Chains:          chains,
		ServiceURL:      serviceURL,
		StakedTokens:    stakedTokens,
		PausedHeight:    uint64(pausedHeight),
		UnstakingHeight: unstakingHeight,
		Output:          output,
	}
	bz, err := cdc.Marshal(&fish)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) UpdateFisherman(address []byte, serviceURL string, amountToAdd string, chains []string) error {
	fish, exists, _ := m.GetFisherman(address)
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	cdc := types.UtilityCodec()
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
	// compute new values
	stakedTokens, err := types.StringToBigInt(fish.StakedTokens)
	if err != nil {
		return err
	}
	stakedTokensToAddI, err := types.StringToBigInt(amountToAdd)
	if err != nil {
		return err
	}
	stakedTokens.Add(stakedTokens, stakedTokensToAddI)
	// update values
	fish.ServiceURL = serviceURL
	fish.StakedTokens = types.BigIntToString(stakedTokens)
	fish.Chains = chains
	// marshal
	bz, err := cdc.Marshal(fish)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) DeleteFisherman(address []byte) error {
	if exists, _ := m.GetFishermanExists(address); !exists {
		return fmt.Errorf("does not exist in world state")
	}
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
	return db.Put(key, DeletedPrefixKey)
}

func (m *MockPersistenceContext) GetFishermanReadyToUnstake(height int64, status int) (Fisherman []modules.UnstakingActor, err error) {
	db := m.Store()
	unstakingKey := append(UnstakingFishermanPrefixKey, []byte(fmt.Sprintf("%d", height))...)
	if has := db.Contains(unstakingKey); !has {
		return nil, nil
	}
	val, err := db.Get(unstakingKey)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return make([]modules.UnstakingActor, 0), nil
	}
	unstakingActors := typespb.UnstakingActors{}
	if err := proto.Unmarshal(val, &unstakingActors); err != nil {
		return nil, err
	}
	for _, sn := range unstakingActors.UnstakingActors {
		Fisherman = append(Fisherman, sn)
	}
	return
}

func (m *MockPersistenceContext) GetFishermanStatus(address []byte) (status int, err error) {
	fish, exists, err := m.GetFisherman(address)
	if err != nil {
		return types.ZeroInt, err
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(fish.Status), nil
}

func (m *MockPersistenceContext) SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	fish, exists, err := m.GetFisherman(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	cdc := types.UtilityCodec()
	unstakingActors := typespb.UnstakingActors{}
	db := m.Store()
	key := append(FishermanPrefixKey, address...)
	fish.UnstakingHeight = unstakingHeight
	fish.Status = int32(status)
	// marshal
	bz, err := cdc.Marshal(fish)
	if err != nil {
		return err
	}
	if err := db.Put(key, bz); err != nil {
		return err
	}
	unstakingKey := append(UnstakingFishermanPrefixKey, []byte(fmt.Sprintf("%d", unstakingHeight))...)
	if found := db.Contains(unstakingKey); found {
		val, err := db.Get(unstakingKey)
		if err != nil {
			return err
		}
		if err := proto.Unmarshal(val, &unstakingActors); err != nil {
			return err
		}
	}
	unstakingActors.UnstakingActors = append(unstakingActors.UnstakingActors, &typespb.UnstakingActor{
		Address:       fish.Address,
		StakeAmount:   fish.StakedTokens,
		OutputAddress: fish.Output,
	})
	unstakingBz, err := cdc.Marshal(&unstakingActors)
	if err != nil {
		return err
	}
	return db.Put(unstakingKey, unstakingBz)
}

func (m *MockPersistenceContext) GetFishermanPauseHeightIfExists(address []byte) (int64, error) {
	fish, exists, err := m.GetFisherman(address)
	if err != nil {
		return types.ZeroInt, nil
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int64(fish.PausedHeight), nil
}

func (m *MockPersistenceContext) SetFishermansStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	db := m.Store()
	cdc := types.UtilityCodec()
	it := db.NewIterator(&util.Range{
		Start: FishermanPrefixKey,
		Limit: PrefixEndBytes(FishermanPrefixKey),
	})
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		fish := typespb.Fisherman{}
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		if err := cdc.Unmarshal(bz, &fish); err != nil {
			return err
		}
		if fish.PausedHeight < uint64(pausedBeforeHeight) {
			fish.UnstakingHeight = unstakingHeight
			fish.Status = int32(status)
			if err := m.SetFishermanUnstakingHeightAndStatus(fish.Address, fish.UnstakingHeight, status); err != nil {
				return err
			}
			bz, err := cdc.Marshal(&fish)
			if err != nil {
				return err
			}
			if err := db.Put(it.Key(), bz); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MockPersistenceContext) SetFishermanPauseHeight(address []byte, height int64) error {
	cdc := types.UtilityCodec()
	db := m.Store()
	fish, exists, err := m.GetFisherman(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	fish.Paused = true
	fish.PausedHeight = uint64(height)
	bz, err := cdc.Marshal(fish)
	if err != nil {
		return err
	}
	return db.Put(append(FishermanPrefixKey, address...), bz)
}

func (m *MockPersistenceContext) GetFishermanOutputAddress(operator []byte) (output []byte, err error) {
	fish, exists, err := m.GetFisherman(operator)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("does not exist in world state")
	}
	return fish.Output, nil
}

func (m *MockPersistenceContext) GetValidator(address []byte) (val *typespb.Validator, exists bool, err error) {
	val = &typespb.Validator{}
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
	bz, err := db.Get(key)
	if err != nil {
		return nil, false, err
	}
	if bz == nil {
		return nil, false, nil
	}
	if bytes.Contains(bz, DeletedPrefixKey) {
		return nil, false, nil
	}
	if err = proto.Unmarshal(bz, val); err != nil {
		return nil, true, err
	}
	return val, true, nil
}

func (m *MockPersistenceContext) GetAllValidators(height int64) (v []*typespb.Validator, err error) {
	cdc := types.UtilityCodec()
	v = make([]*typespb.Validator, 0)
	var it iterator.Iterator
	if height == m.Height {
		db := m.Store()
		it = db.NewIterator(&util.Range{
			Start: ValidatorPrefixKey,
			Limit: PrefixEndBytes(ValidatorPrefixKey),
		})
	} else {
		key := HeightKey(height, ValidatorPrefixKey)
		it = m.Parent.GetCommitDB().NewIterator(&util.Range{
			Start: key,
			Limit: PrefixEndBytes(key),
		})
	}
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		bz := it.Value()
		//if bz == nil {
		//	break
		//}
		valid := it.Valid()
		valid = valid
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		validator := typespb.Validator{}
		if err := cdc.Unmarshal(bz, &validator); err != nil {
			return nil, err
		}
		v = append(v, &validator)
	}
	return
}

func (m *MockPersistenceContext) GetValidatorExists(address []byte) (exists bool, err error) {
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
	if found := db.Contains(key); !found {
		return false, nil
	}
	bz, err := db.Get(key)
	if err != nil {
		return false, err
	}
	if bz == nil {
		return false, nil
	}
	if bytes.Contains(bz, DeletedPrefixKey) {
		return false, nil
	}
	return true, nil
}

func (m *MockPersistenceContext) InsertValidator(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, pausedHeight int64, unstakingHeight int64) error {
	if _, exists, _ := m.GetFisherman(address); exists {
		return fmt.Errorf("already exists in world state")
	}
	cdc := types.UtilityCodec()
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
	val := typespb.Validator{
		Address:         address,
		PublicKey:       publicKey,
		Paused:          paused,
		Status:          int32(status),
		ServiceURL:      serviceURL,
		StakedTokens:    stakedTokens,
		MissedBlocks:    0,
		PausedHeight:    uint64(pausedHeight),
		UnstakingHeight: unstakingHeight,
		Output:          output,
	}
	bz, err := cdc.Marshal(&val)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) UpdateValidator(address []byte, serviceURL string, amountToAdd string) error {
	val, exists, _ := m.GetValidator(address)
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	cdc := types.UtilityCodec()
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
	// compute new values
	stakedTokens, err := types.StringToBigInt(val.StakedTokens)
	if err != nil {
		return err
	}
	stakedTokensToAddI, err := types.StringToBigInt(amountToAdd)
	if err != nil {
		return err
	}
	stakedTokens.Add(stakedTokens, stakedTokensToAddI)
	// update values
	val.ServiceURL = serviceURL
	val.StakedTokens = types.BigIntToString(stakedTokens)
	// marshal
	bz, err := cdc.Marshal(val)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *MockPersistenceContext) DeleteValidator(address []byte) error {
	if exists, _ := m.GetValidatorExists(address); !exists {
		return fmt.Errorf("does not exist in world state")
	}
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
	return db.Put(key, DeletedPrefixKey)
}

func (m *MockPersistenceContext) GetValidatorsReadyToUnstake(height int64, status int) (fishermen []modules.UnstakingActor, err error) {
	db := m.Store()
	unstakingKey := append(UnstakingValidatorPrefixKey, []byte(fmt.Sprintf("%d", height))...)
	if has := db.Contains(unstakingKey); !has {
		return nil, nil
	}
	val, err := db.Get(unstakingKey)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return make([]modules.UnstakingActor, 0), nil
	}
	unstakingActors := typespb.UnstakingActors{}
	if err := proto.Unmarshal(val, &unstakingActors); err != nil {
		return nil, err
	}
	for _, sn := range unstakingActors.UnstakingActors {
		fishermen = append(fishermen, sn)
	}
	return
}

func (m *MockPersistenceContext) GetValidatorStatus(address []byte) (status int, err error) {
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return types.ZeroInt, err
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(val.Status), nil
}

func (m *MockPersistenceContext) SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error {
	validator, exists, err := m.GetValidator(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	cdc := types.UtilityCodec()
	unstakingActors := typespb.UnstakingActors{}
	db := m.Store()
	key := append(ValidatorPrefixKey, address...)
	validator.UnstakingHeight = unstakingHeight
	validator.Status = int32(status)
	// marshal
	bz, err := cdc.Marshal(validator)
	if err != nil {
		return err
	}
	if err := db.Put(key, bz); err != nil {
		return err
	}
	unstakingKey := append(UnstakingValidatorPrefixKey, []byte(fmt.Sprintf("%d", unstakingHeight))...)
	if found := db.Contains(unstakingKey); found {
		val, err := db.Get(unstakingKey)
		if err != nil {
			return err
		}
		if err := proto.Unmarshal(val, &unstakingActors); err != nil {
			return err
		}
	}
	unstakingActors.UnstakingActors = append(unstakingActors.UnstakingActors, &typespb.UnstakingActor{
		Address:       validator.Address,
		StakeAmount:   validator.StakedTokens,
		OutputAddress: validator.Output,
	})
	unstakingBz, err := cdc.Marshal(&unstakingActors)
	if err != nil {
		return err
	}
	return db.Put(unstakingKey, unstakingBz)
}

func (m *MockPersistenceContext) GetValidatorPauseHeightIfExists(address []byte) (int64, error) {
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return types.ZeroInt, nil
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int64(val.PausedHeight), nil
}

func (m *MockPersistenceContext) SetValidatorsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error {
	db := m.Store()
	cdc := types.UtilityCodec()
	it := db.NewIterator(&util.Range{
		Start: ValidatorPrefixKey,
		Limit: PrefixEndBytes(ValidatorPrefixKey),
	})
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		validator := typespb.Validator{}
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		if err := cdc.Unmarshal(bz, &validator); err != nil {
			return err
		}
		if validator.PausedHeight < uint64(pausedBeforeHeight) {
			validator.UnstakingHeight = unstakingHeight
			validator.Status = int32(status)
			if err := m.SetFishermanUnstakingHeightAndStatus(validator.Address, validator.UnstakingHeight, status); err != nil {
				return err
			}
			bz, err := cdc.Marshal(&validator)
			if err != nil {
				return err
			}
			if err := db.Put(it.Key(), bz); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MockPersistenceContext) SetValidatorPauseHeightAndMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) error {
	cdc := types.UtilityCodec()
	db := m.Store()
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	val.PausedHeight = uint64(pauseHeight)
	val.MissedBlocks = uint32(missedBlocks)
	bz, err := cdc.Marshal(val)
	if err != nil {
		return err
	}
	return db.Put(append(ValidatorPrefixKey, address...), bz)
}

func (m *MockPersistenceContext) GetValidatorMissedBlocks(address []byte) (int, error) {
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return types.ZeroInt, err
	}
	if !exists {
		return types.ZeroInt, fmt.Errorf("does not exist in world state")
	}
	return int(val.MissedBlocks), nil
}

func (m *MockPersistenceContext) SetValidatorPauseHeight(address []byte, height int64) error {
	cdc := types.UtilityCodec()
	db := m.Store()
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	val.Paused = true
	val.PausedHeight = uint64(height)
	bz, err := cdc.Marshal(val)
	if err != nil {
		return err
	}
	return db.Put(append(ValidatorPrefixKey, address...), bz)
}

func (m *MockPersistenceContext) SetValidatorStakedTokens(address []byte, tokens string) error {
	cdc := types.UtilityCodec()
	db := m.Store()
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("does not exist in world state")
	}
	val.StakedTokens = tokens
	bz, err := cdc.Marshal(val)
	if err != nil {
		return err
	}
	return db.Put(append(ValidatorPrefixKey, address...), bz)
}

func (m *MockPersistenceContext) GetValidatorStakedTokens(address []byte) (tokens string, err error) {
	val, exists, err := m.GetValidator(address)
	if err != nil {
		return types.EmptyString, err
	}
	if !exists {
		return types.EmptyString, fmt.Errorf("does not exist in world state")
	}
	return val.StakedTokens, nil
}

func (m *MockPersistenceContext) GetValidatorOutputAddress(operator []byte) (output []byte, err error) {
	val, exists, err := m.GetValidator(operator)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("does not exist in world state")
	}
	return val.Output, nil
}

func (m *MockPersistenceContext) GetBlocksPerSession() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.BlocksPerSession), nil
}

func (m *MockPersistenceContext) GetParamAppMinimumStake() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.GetAppMinimumStake(), nil
}

func (m *MockPersistenceContext) GetMaxAppChains() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.AppMaxChains), nil
}

func (m *MockPersistenceContext) GetBaselineAppStakeRate() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.AppBaselineStakeRate), nil
}

func (m *MockPersistenceContext) GetStakingAdjustment() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.AppStakingAdjustment), nil
}

func (m *MockPersistenceContext) GetAppUnstakingBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.AppUnstakingBlocks), nil
}

func (m *MockPersistenceContext) GetAppMinimumPauseBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.AppMinimumPauseBlocks), nil
}

func (m *MockPersistenceContext) GetAppMaxPausedBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.AppMaxPauseBlocks), nil
}

func (m *MockPersistenceContext) GetParamServiceNodeMinimumStake() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.ServiceNodeMinimumStake, nil
}

func (m *MockPersistenceContext) GetServiceNodeMaxChains() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ServiceNodeMaxChains), nil
}

func (m *MockPersistenceContext) GetServiceNodeUnstakingBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ServiceNodeUnstakingBlocks), nil
}

func (m *MockPersistenceContext) GetServiceNodeMinimumPauseBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ServiceNodeMinimumPauseBlocks), nil
}

func (m *MockPersistenceContext) GetServiceNodeMaxPausedBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ServiceNodeMaxPauseBlocks), nil
}

func (m *MockPersistenceContext) GetServiceNodesPerSession() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ServiceNodesPerSession), nil
}

func (m *MockPersistenceContext) GetParamFishermanMinimumStake() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.FishermanMinimumStake, nil
}

func (m *MockPersistenceContext) GetFishermanMaxChains() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.FishermanMaxChains), nil
}

func (m *MockPersistenceContext) GetFishermanUnstakingBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.FishermanUnstakingBlocks), nil
}

func (m *MockPersistenceContext) GetFishermanMinimumPauseBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.FishermanMinimumPauseBlocks), nil
}

func (m *MockPersistenceContext) GetFishermanMaxPausedBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.FishermanMaxPauseBlocks), nil
}

func (m *MockPersistenceContext) GetParamValidatorMinimumStake() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.ValidatorMinimumStake, nil
}

func (m *MockPersistenceContext) GetValidatorUnstakingBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ValidatorUnstakingBlocks), nil
}

func (m *MockPersistenceContext) GetValidatorMinimumPauseBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ValidatorMinimumPauseBlocks), nil
}

func (m *MockPersistenceContext) GetValidatorMaxPausedBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ValidatorMaxPauseBlocks), nil
}

func (m *MockPersistenceContext) GetValidatorMaximumMissedBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ValidatorMaximumMissedBlocks), nil
}

func (m *MockPersistenceContext) GetProposerPercentageOfFees() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ProposerPercentageOfFees), nil
}

func (m *MockPersistenceContext) GetMaxEvidenceAgeInBlocks() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.ValidatorMaxEvidenceAgeInBlocks), nil
}

func (m *MockPersistenceContext) GetMissedBlocksBurnPercentage() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.MissedBlocksBurnPercentage), nil
}

func (m *MockPersistenceContext) GetDoubleSignBurnPercentage() (int, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.DoubleSignBurnPercentage), nil
}

func (m *MockPersistenceContext) GetMessageDoubleSignFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageDoubleSignFee, nil
}

func (m *MockPersistenceContext) GetMessageSendFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageSendFee, nil
}

func (m *MockPersistenceContext) GetMessageStakeFishermanFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageStakeFishermanFee, nil
}

func (m *MockPersistenceContext) GetMessageEditStakeFishermanFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageEditStakeFishermanFee, nil
}

func (m *MockPersistenceContext) GetMessageUnstakeFishermanFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageUnstakeFishermanFee, nil
}

func (m *MockPersistenceContext) GetMessagePauseFishermanFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessagePauseFishermanFee, nil
}

func (m *MockPersistenceContext) GetMessageUnpauseFishermanFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageUnpauseFishermanFee, nil
}

func (m *MockPersistenceContext) GetMessageFishermanPauseServiceNodeFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessagePauseServiceNodeFee, nil
}

func (m *MockPersistenceContext) GetMessageTestScoreFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageProveTestScoreFee, nil
}

func (m *MockPersistenceContext) GetMessageProveTestScoreFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageProveTestScoreFee, nil
}

func (m *MockPersistenceContext) GetMessageStakeAppFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageStakeAppFee, nil
}

func (m *MockPersistenceContext) GetMessageEditStakeAppFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageEditStakeAppFee, nil
}

func (m *MockPersistenceContext) GetMessageUnstakeAppFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageUnstakeAppFee, nil
}

func (m *MockPersistenceContext) GetMessagePauseAppFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessagePauseAppFee, nil
}

func (m *MockPersistenceContext) GetMessageUnpauseAppFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageUnpauseAppFee, nil
}

func (m *MockPersistenceContext) GetMessageStakeValidatorFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageStakeValidatorFee, nil
}

func (m *MockPersistenceContext) GetMessageEditStakeValidatorFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageEditStakeValidatorFee, nil
}

func (m *MockPersistenceContext) GetMessageUnstakeValidatorFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageUnstakeValidatorFee, nil
}

func (m *MockPersistenceContext) GetMessagePauseValidatorFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessagePauseValidatorFee, nil
}

func (m *MockPersistenceContext) GetMessageUnpauseValidatorFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageUnpauseValidatorFee, nil
}

func (m *MockPersistenceContext) GetMessageStakeServiceNodeFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageStakeServiceNodeFee, nil
}

func (m *MockPersistenceContext) GetMessageEditStakeServiceNodeFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageEditStakeServiceNodeFee, nil
}

func (m *MockPersistenceContext) GetMessageUnstakeServiceNodeFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageUnstakeServiceNodeFee, nil
}

func (m *MockPersistenceContext) GetMessagePauseServiceNodeFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessagePauseServiceNodeFee, nil
}

func (m *MockPersistenceContext) GetMessageUnpauseServiceNodeFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageUnpauseServiceNodeFee, nil
}

func (m *MockPersistenceContext) GetMessageChangeParameterFee() (string, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return types.EmptyString, err
	}
	return params.MessageChangeParameterFee, nil
}

func (m *MockPersistenceContext) SetParams(p *typespb.Params) error {
	cdc := types.UtilityCodec()
	store := m.Store()
	bz, err := cdc.Marshal(p)
	if err != nil {
		return err
	}
	return store.Put(ParamsPrefixKey, bz)
}

func (m *MockPersistenceContext) SetBlocksPerSession(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.BlocksPerSession = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetParamAppMinimumStake(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMinimumStake = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMaxAppChains(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMaxChains = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetBaselineAppStakeRate(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppBaselineStakeRate = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetStakingAdjustment(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppStakingAdjustment = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetAppUnstakingBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppUnstakingBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetAppMinimumPauseBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMinimumPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetAppMaxPausedBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMaxPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetParamServiceNodeMinimumStake(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMinimumStake = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetServiceNodeMaxChains(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMaxChains = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetServiceNodeUnstakingBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeUnstakingBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetServiceNodeMinimumPauseBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMinimumPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetServiceNodeMaxPausedBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMaxPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetServiceNodesPerSession(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodesPerSession = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetParamFishermanMinimumStake(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMinimumStake = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetFishermanMaxChains(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMaxChains = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetFishermanUnstakingBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanUnstakingBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetFishermanMinimumPauseBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMinimumPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetFishermanMaxPausedBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMaxPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetParamValidatorMinimumStake(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMinimumStake = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetValidatorUnstakingBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorUnstakingBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetValidatorMinimumPauseBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMinimumPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetValidatorMaxPausedBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaxPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetValidatorMaximumMissedBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaximumMissedBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetProposerPercentageOfFees(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ProposerPercentageOfFees = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMaxEvidenceAgeInBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaxEvidenceAgeInBlocks = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMissedBlocksBurnPercentage(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MissedBlocksBurnPercentage = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetDoubleSignBurnPercentage(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.DoubleSignBurnPercentage = int32(i)
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageDoubleSignFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageDoubleSignFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageSendFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageSendFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageStakeFishermanFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeFishermanFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageEditStakeFishermanFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeFishermanFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnstakeFishermanFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeFishermanFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessagePauseFishermanFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseFishermanFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnpauseFishermanFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseFishermanFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageFishermanPauseServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseServiceNodeFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageTestScoreFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageTestScoreFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageProveTestScoreFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageProveTestScoreFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageStakeAppFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeAppFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageEditStakeAppFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeAppFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnstakeAppFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeAppFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessagePauseAppFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseAppFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnpauseAppFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseAppFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageStakeValidatorFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeValidatorFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageEditStakeValidatorFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeValidatorFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnstakeValidatorFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeValidatorFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessagePauseValidatorFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseValidatorFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnpauseValidatorFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseValidatorFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageStakeServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeServiceNodeFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageEditStakeServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeServiceNodeFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnstakeServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeServiceNodeFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessagePauseServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageFishermanPauseServiceNodeFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnpauseServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseServiceNodeFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageChangeParameterFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageChangeParameterFee = s
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageDoubleSignFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageDoubleSignFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageSendFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageSendFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageStakeFishermanFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeFishermanFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageEditStakeFishermanFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeFishermanFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnstakeFishermanFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeFishermanFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessagePauseFishermanFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseFishermanFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnpauseFishermanFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseFishermanFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageFishermanPauseServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageFishermanPauseServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageTestScoreFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageTestScoreFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageProveTestScoreFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageProveTestScoreFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageStakeAppFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeAppFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageEditStakeAppFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeAppFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnstakeAppFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeAppFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessagePauseAppFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseAppFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnpauseAppFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseAppFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageStakeValidatorFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeValidatorFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageEditStakeValidatorFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeValidatorFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnstakeValidatorFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeValidatorFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessagePauseValidatorFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseValidatorFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnpauseValidatorFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseValidatorFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageStakeServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageEditStakeServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnstakeServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessagePauseServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageUnpauseServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetMessageChangeParameterFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageChangeParameterFeeOwner = bytes
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetACLOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ACLOwner, nil
}

func (m *MockPersistenceContext) SetACLOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ACLOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetBlocksPerSessionOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.BlocksPerSessionOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetBlocksPerSessionOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.BlocksPerSessionOwner, nil
}

func (m *MockPersistenceContext) GetMaxAppChainsOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.AppMaxChainsOwner, nil
}

func (m *MockPersistenceContext) SetMaxAppChainsOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMaxChainsOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetAppMinimumStakeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.AppMinimumStakeOwner, nil
}

func (m *MockPersistenceContext) SetAppMinimumStakeOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMinimumStakeOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetBaselineAppOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.AppBaselineStakeRateOwner, nil
}

func (m *MockPersistenceContext) SetBaselineAppOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppBaselineStakeRateOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetStakingAdjustmentOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.AppStakingAdjustmentOwner, nil
}

func (m *MockPersistenceContext) SetStakingAdjustmentOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppStakingAdjustmentOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetAppUnstakingBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.AppUnstakingBlocksOwner, nil
}

func (m *MockPersistenceContext) SetAppUnstakingBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppUnstakingBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetAppMinimumPauseBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.AppMinimumPauseBlocksOwner, nil
}

func (m *MockPersistenceContext) SetAppMinimumPauseBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMinimumPauseBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetAppMaxPausedBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.AppMaxPausedBlocksOwner, nil
}

func (m *MockPersistenceContext) SetAppMaxPausedBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMaxPausedBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetParamServiceNodeMinimumStakeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ServiceNodeMinimumStakeOwner, nil
}

func (m *MockPersistenceContext) SetParamServiceNodeMinimumStakeOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMinimumStakeOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetServiceNodeMaxChainsOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ServiceNodeMaxChainsOwner, nil
}

func (m *MockPersistenceContext) SetMaxServiceNodeChainsOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMaxChainsOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetServiceNodeUnstakingBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ServiceNodeUnstakingBlocksOwner, nil
}

func (m *MockPersistenceContext) SetServiceNodeUnstakingBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeUnstakingBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetServiceNodeMinimumPauseBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ServiceNodeMinimumPauseBlocksOwner, nil
}

func (m *MockPersistenceContext) SetServiceNodeMinimumPauseBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMinimumStakeOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetServiceNodeMaxPausedBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ServiceNodeMaxPausedBlocksOwner, nil
}

func (m *MockPersistenceContext) SetServiceNodeMaxPausedBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMaxPausedBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetFishermanMinimumStakeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ParamFishermanMinimumStakeOwner, nil
}

func (m *MockPersistenceContext) SetFishermanMinimumStakeOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ParamFishermanMinimumStakeOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetMaxFishermanChainsOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.FishermanMaxChainsOwner, nil
}

func (m *MockPersistenceContext) SetMaxFishermanChainsOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMaxChainsOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetFishermanUnstakingBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.FishermanUnstakingBlocksOwner, nil
}

func (m *MockPersistenceContext) SetFishermanUnstakingBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanUnstakingBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetFishermanMinimumPauseBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.FishermanMinimumPauseBlocksOwner, nil
}

func (m *MockPersistenceContext) SetFishermanMinimumPauseBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMinimumPauseBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetFishermanMaxPausedBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.FishermanMaxPausedBlocksOwner, nil
}

func (m *MockPersistenceContext) SetFishermanMaxPausedBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMaxPausedBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetParamValidatorMinimumStakeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ValidatorMinimumStakeOwner, nil
}

func (m *MockPersistenceContext) SetParamValidatorMinimumStakeOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMinimumStakeOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetValidatorUnstakingBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ValidatorUnstakingBlocksOwner, nil
}

func (m *MockPersistenceContext) SetValidatorUnstakingBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorUnstakingBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetValidatorMinimumPauseBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ValidatorMinimumPauseBlocksOwner, nil
}

func (m *MockPersistenceContext) SetValidatorMinimumPauseBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMinimumPauseBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetValidatorMaxPausedBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ValidatorMaxPausedBlocksOwner, nil
}

func (m *MockPersistenceContext) SetValidatorMaxPausedBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaxPausedBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetValidatorMaximumMissedBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ValidatorMaximumMissedBlocksOwner, nil
}

func (m *MockPersistenceContext) SetValidatorMaximumMissedBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaximumMissedBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetProposerPercentageOfFeesOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ProposerPercentageOfFeesOwner, nil
}

func (m *MockPersistenceContext) SetProposerPercentageOfFeesOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ProposerPercentageOfFeesOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetMaxEvidenceAgeInBlocksOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ValidatorMaxEvidenceAgeInBlocksOwner, nil
}

func (m *MockPersistenceContext) SetMaxEvidenceAgeInBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaxEvidenceAgeInBlocksOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetMissedBlocksBurnPercentageOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MissedBlocksBurnPercentageOwner, nil
}

func (m *MockPersistenceContext) SetMissedBlocksBurnPercentageOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MissedBlocksBurnPercentageOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetDoubleSignBurnPercentageOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.DoubleSignBurnPercentageOwner, nil
}

func (m *MockPersistenceContext) SetDoubleSignBurnPercentageOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.DoubleSignBurnPercentageOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) SetServiceNodesPerSessionOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodesPerSessionOwner = owner
	return m.SetParams(params)
}

func (m *MockPersistenceContext) GetServiceNodesPerSessionOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.ServiceNodesPerSessionOwner, nil
}

func (m *MockPersistenceContext) GetMessageDoubleSignFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageDoubleSignFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageSendFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageSendFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageStakeFishermanFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageStakeFishermanFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageEditStakeFishermanFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageEditStakeFishermanFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageUnstakeFishermanFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageUnstakeFishermanFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessagePauseFishermanFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessagePauseFishermanFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageUnpauseFishermanFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageUnpauseFishermanFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageFishermanPauseServiceNodeFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageFishermanPauseServiceNodeFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageTestScoreFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageTestScoreFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageProveTestScoreFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageProveTestScoreFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageStakeAppFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageEditStakeAppFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageEditStakeAppFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageEditStakeAppFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageUnstakeAppFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageUnstakeAppFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessagePauseAppFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessagePauseAppFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageUnpauseAppFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageUnpauseAppFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageStakeValidatorFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageStakeValidatorFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageEditStakeValidatorFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageEditStakeValidatorFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageUnstakeValidatorFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageUnstakeValidatorFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessagePauseValidatorFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessagePauseValidatorFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageUnpauseValidatorFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageUnpauseValidatorFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageStakeServiceNodeFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageStakeServiceNodeFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageEditStakeServiceNodeFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageEditStakeServiceNodeFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageUnstakeServiceNodeFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageUnstakeServiceNodeFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessagePauseServiceNodeFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessagePauseServiceNodeFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageUnpauseServiceNodeFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageUnpauseServiceNodeFeeOwner, nil
}

func (m *MockPersistenceContext) GetMessageChangeParameterFeeOwner() ([]byte, error) {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return nil, err
	}
	return params.MessageChangeParameterFeeOwner, nil
}

func NewMemDB() *memdb.DB {
	return memdb.New(comparer.DefaultComparer, 100000)
}

func CopyMemDB(src, dest *memdb.DB) error {
	it := src.NewIterator(&util.Range{})
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		err := dest.Put(it.Key(), it.Value())
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	elenEncoder = lexnum.NewEncoder('=', '-')
)

func HeightKey(height int64, k []byte) (key []byte) {
	keyString := fmt.Sprintf("%s/%s", elenEncoder.EncodeInt(int(height)), k)
	return []byte(keyString)
}

func KeyFromHeightKey(heightKey []byte) (key []byte) {
	k := strings.SplitN(string(heightKey), "/", 2)[1]
	return []byte(k)
}

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

func ChainsEquality(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
