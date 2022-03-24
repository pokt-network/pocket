package pre_persistence

import (
	"bytes"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"

	"math/big"

	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	ServiceNodeStakePoolName = "SERVICE_NODE_STAKE_POOL"
	AppStakePoolName         = "APP_STAKE_POOL"
	ValidatorStakePoolName   = "VALIDATOR_STAKE_POOL"
	FishermanStakePoolName   = "FISHERMAN_STAKE_POOL"
	DAOPoolName              = "DAO_POOL"
	FeePoolName              = "FEE_POOL"
)

func (m *PrePersistenceContext) AddPoolAmount(name string, amount string) error {
	cdc := Cdc()
	p := Pool{}
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
	s, err := StringToBigInt(p.Account.Amount)
	if err != nil {
		return err
	}
	s2, err := StringToBigInt(amount)
	if err != nil {
		return err
	}
	s.Add(s, s2)
	p.Account.Amount = BigIntToString(s)
	bz, err := cdc.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) SubtractPoolAmount(name string, amount string) error {
	cdc := Cdc()
	p := Pool{}
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
	s, err := StringToBigInt(p.Account.Amount)
	if err != nil {
		return err
	}
	s2, err := StringToBigInt(amount)
	if err != nil {
		return err
	}
	s.Sub(s, s2)
	p.Account.Amount = BigIntToString(s)
	bz, err := cdc.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) InsertPool(name string, address []byte, amount string) error {
	cdc := Cdc()
	p := Pool{
		Name: name,
		Account: &Account{
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

func (m *PrePersistenceContext) SetPoolAmount(name string, amount string) error {
	cdc := Cdc()
	p := Pool{}
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

func (m *PrePersistenceContext) GetPoolAmount(name string) (amount string, err error) {
	cdc := Cdc()
	p := Pool{}
	db := m.Store()
	key := append(PoolPrefixKey, []byte(name)...)
	val, err := db.Get(key)
	if err != nil {
		return EmptyString, err
	}
	err = cdc.Unmarshal(val, &p)
	if err != nil {
		return EmptyString, err
	}
	return p.Account.Amount, nil
}

func (m *PrePersistenceContext) GetAllPools(height int64) (pools []*Pool, err error) {
	cdc := Cdc()
	pools = make([]*Pool, 0)
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
		p := Pool{}
		if err := cdc.Unmarshal(bz, &p); err != nil {
			return nil, err
		}
		pools = append(pools, &p)
	}
	return
}

func (m *PrePersistenceContext) AddAccountAmount(address []byte, amount string) error {
	cdc := Cdc()
	account := Account{
		Amount: BigIntToString(big.NewInt(0)),
	}
	db := m.Store()
	key := append(AccountPrefixKey, address...)
	if db.Contains(key) {
		val, err := db.Get(key)
		if err != nil {
			return err
		}
		err = cdc.Unmarshal(val, &account)
		if err != nil {
			return err
		}
	}
	s, err := StringToBigInt(account.Amount)
	if err != nil {
		return err
	}
	s2, err := StringToBigInt(amount)
	if err != nil {
		return err
	}
	s.Add(s, s2)
	account.Amount = BigIntToString(s)
	bz, err := cdc.Marshal(&account)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) SubtractAccountAmount(address []byte, amount string) error {
	cdc := Cdc()
	account := Account{}
	db := m.Store()
	key := append(AccountPrefixKey, address...)
	val, err := db.Get(key)
	if err != nil {
		return err
	}
	err = cdc.Unmarshal(val, &account)
	if err != nil {
		return err
	}
	s, err := StringToBigInt(account.Amount)
	if err != nil {
		return err
	}
	s2, err := StringToBigInt(amount)
	if err != nil {
		return err
	}
	s.Sub(s, s2)
	account.Amount = BigIntToString(s)
	bz, err := cdc.Marshal(&account)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) GetAccountAmount(address []byte) (string, error) {
	cdc := Cdc()
	account := Account{}
	db := m.Store()
	key := append(AccountPrefixKey, address...)
	val, err := db.Get(key)
	if err != nil {
		return EmptyString, err
	}
	err = cdc.Unmarshal(val, &account)
	if err != nil {
		return EmptyString, err
	}
	return account.Amount, nil
}

func (m *PrePersistenceContext) SetAccount(address []byte, amount string) error {
	cdc := Cdc()
	account := Account{
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

func (m *PrePersistenceContext) GetAllAccounts(height int64) (accs []*Account, err error) {
	cdc := Cdc()
	accs = make([]*Account, 0)
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
		acc := Account{}
		if err := cdc.Unmarshal(bz, &acc); err != nil {
			return nil, err
		}
		accs = append(accs, &acc)
	}
	return
}

func (x *Account) ValidateBasic() types.Error {
	if x == nil {
		return types.ErrEmptyAccount()
	}
	if x.Address == nil {
		return types.ErrEmptyAddress()
	}
	addrLen := len(x.Address)
	if addrLen != crypto.AddressLen {
		return types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen(addrLen))
	}
	amount := big.NewInt(0)
	if _, ok := amount.SetString(x.Amount, 10); !ok {
		return types.ErrInvalidAmount()
	}
	return nil
}

func (x *Account) SetAddress(address crypto.Address) types.Error {
	if x == nil {
		return types.ErrEmptyAccount()
	}
	addrLen := len(address)
	if addrLen != crypto.AddressLen {
		return types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen(addrLen))
	}
	x.Address = address
	return nil
}

func (x *Account) SetAmount(amount big.Int) types.Error {
	if x == nil {
		return types.ErrEmptyAccount()
	}
	x.Amount = amount.String()
	return nil
}

func NewPool(name string, account *Account) (*Pool, types.Error) {
	pool := &Pool{}
	if err := pool.SetName(name); err != nil {
		return nil, err
	}
	if err := pool.SetAccount(account); err != nil {
		return nil, err
	}
	return pool, nil
}

func (x *Pool) ValidateBasic() types.Error {
	if x == nil {
		return types.ErrNilPool()
	}
	if x.Name == "" {
		return types.ErrEmptyName()
	}
	return x.Account.ValidateBasic()
}

func (x *Pool) SetName(name string) types.Error {
	if name == "" {
		return types.ErrEmptyName()
	}
	if x == nil {
		return types.ErrNilPool()
	}
	x.Name = name
	return nil
}

func (x *Pool) SetAccount(account *Account) types.Error {
	if x == nil {
		return types.ErrNilPool()
	}
	if account == nil {
		return types.ErrEmptyAccount()
	}
	if err := account.ValidateBasic(); err != nil {
		return err
	}
	x.Account = account
	return nil
}
