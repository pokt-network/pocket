package pre_persistence

import (
	"bytes"
	"math/big"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const ( // Names for each 'pool' (specialized accounts)
	ServiceNodeStakePoolName = "SERVICE_NODE_STAKE_POOL"
	AppStakePoolName         = "APP_STAKE_POOL"
	ValidatorStakePoolName   = "VALIDATOR_STAKE_POOL"
	FishermanStakePoolName   = "FISHERMAN_STAKE_POOL"
	DAOPoolName              = "DAO_POOL"
	FeePoolName              = "FEE_POOL"
)

func (m *PrePersistenceContext) AddPoolAmount(name string, amount string) error {
	add := func(s *big.Int, s1 *big.Int) error {
		s.Add(s, s1)
		return nil
	}
	return m.operationPoolAmount(name, amount, add)
}

func (m *PrePersistenceContext) SubtractPoolAmount(name string, amount string) error {
	sub := func(s *big.Int, s1 *big.Int) error {
		s.Sub(s, s1)
		if s.Sign() == -1 {
			return types.ErrInsufficientAmountError()
		}
		return nil
	}
	return m.operationPoolAmount(name, amount, sub)
}

func (m *PrePersistenceContext) operationPoolAmount(name string, amount string, op func(*big.Int, *big.Int) error) error {
	codec := GetCodec()
	p := Pool{}
	db := m.Store()
	key := append(PoolPrefixKey, []byte(name)...)
	val, err := db.Get(key)
	if err != nil {
		return err
	}
	err = codec.Unmarshal(val, &p)
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
	if err := op(s, s2); err != nil {
		return err
	}
	p.Account.Amount = BigIntToString(s)
	bz, err := codec.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) InsertPool(name string, address []byte, amount string) error {
	codec := GetCodec()
	p := Pool{
		Name: name,
		Account: &Account{
			Address: address,
			Amount:  amount,
		},
	}
	db := m.Store()
	key := append(PoolPrefixKey, []byte(name)...)
	bz, err := codec.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) SetPoolAmount(name string, amount string) error {
	codec := GetCodec()
	p := Pool{}
	db := m.Store()
	key := append(PoolPrefixKey, []byte(name)...)
	val, err := db.Get(key)
	if err != nil {
		return err
	}
	err = codec.Unmarshal(val, &p)
	if err != nil {
		return err
	}
	p.Account.Amount = amount
	bz, err := codec.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) GetPoolAmount(name string) (amount string, err error) {
	codec := GetCodec()
	p := Pool{}
	db := m.Store()
	key := append(PoolPrefixKey, []byte(name)...)
	val, err := db.Get(key)
	if err != nil {
		return EmptyString, err
	}
	err = codec.Unmarshal(val, &p)
	if err != nil {
		return EmptyString, err
	}
	return p.Account.Amount, nil
}

func (m *PrePersistenceContext) GetAllPools(height int64) (pools []*Pool, err error) {
	codec := GetCodec()
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
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		p := Pool{}
		if err := codec.Unmarshal(bz, &p); err != nil {
			return nil, err
		}
		pools = append(pools, &p)
	}
	return
}

func (m *PrePersistenceContext) operationAccountAmount(address []byte, amount string, op func(*big.Int, *big.Int) error) error {
	codec := GetCodec()
	a := Account{}
	db := m.Store()
	key := append(AccountPrefixKey, address...)
	val, err := db.Get(key)
	if err != nil {
		return err
	}
	err = codec.Unmarshal(val, &a)
	if err != nil {
		return err
	}
	s, err := StringToBigInt(a.Amount)
	if err != nil {
		return err
	}
	s2, err := StringToBigInt(amount)
	if err != nil {
		return err
	}
	if err := op(s, s2); err != nil {
		return err
	}
	a.Amount = BigIntToString(s)
	bz, err := codec.Marshal(&a)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) AddAccountAmount(address []byte, amount string) error {
	add := func(s *big.Int, s1 *big.Int) error {
		s.Add(s, s1)
		return nil
	}
	return m.operationAccountAmount(address, amount, add)
}

func (m *PrePersistenceContext) SubtractAccountAmount(address []byte, amount string) error {
	sub := func(s *big.Int, s1 *big.Int) error {
		s.Sub(s, s1)
		if s.Sign() == -1 {
			return types.ErrInsufficientAmountError()
		}
		return nil
	}
	return m.operationAccountAmount(address, amount, sub)
}

func (m *PrePersistenceContext) GetAccountAmount(address []byte) (string, error) {
	codec := GetCodec()
	account := Account{}
	db := m.Store()
	key := append(AccountPrefixKey, address...)
	val, err := db.Get(key)
	if err != nil {
		return EmptyString, err
	}
	err = codec.Unmarshal(val, &account)
	if err != nil {
		return EmptyString, err
	}
	return account.Amount, nil
}

func (m *PrePersistenceContext) SetAccount(address []byte, amount string) error {
	codec := GetCodec()
	account := Account{
		Address: address,
		Amount:  amount,
	}
	db := m.Store()
	key := append(AccountPrefixKey, address...)
	bz, err := codec.Marshal(&account)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) GetAllAccounts(height int64) (accs []*Account, err error) {
	codec := GetCodec()
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
	defer it.Release()
	for valid := it.First(); valid; valid = it.Next() {
		bz := it.Value()
		if bytes.Contains(bz, DeletedPrefixKey) {
			continue
		}
		acc := Account{}
		if err := codec.Unmarshal(bz, &acc); err != nil {
			return nil, err
		}
		accs = append(accs, &acc)
	}
	return
}

func (a *Account) ValidateBasic() types.Error {
	if a == nil {
		return types.ErrEmptyAccount()
	}
	if a.Address == nil {
		return types.ErrEmptyAddress()
	}
	if len(a.Address) != crypto.AddressLen {
		return types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen())
	}
	amount := &big.Int{}
	if _, ok := amount.SetString(a.Amount, 10); !ok {
		return types.ErrInvalidAmount()
	}
	return nil
}

func (a *Account) SetAddress(address crypto.Address) types.Error {
	if a == nil {
		return types.ErrEmptyAccount()
	}
	if len(address) != crypto.AddressLen {
		return types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen())
	}
	a.Address = address
	return nil
}

func (a *Account) SetAccountAmount(amount big.Int) types.Error {
	if a == nil {
		return types.ErrEmptyAccount()
	}
	if amount.Sign() == -1 {
		return types.ErrNegativeAmountError()
	}
	a.Amount = amount.String()
	return nil
}

func NewPool(name string, account *Account) (*Pool, types.Error) {
	pool := &Pool{
		Name:    name,
		Account: account,
	}
	if err := pool.ValidateBasic(); err != nil {
		return nil, err
	}
	return pool, nil
}

func (p *Pool) ValidateBasic() types.Error {
	if p == nil {
		return types.ErrNilPool()
	}
	if p.Name == "" {
		return types.ErrEmptyName()
	}
	return p.Account.ValidateBasic()
}

func (p *Pool) SetName(name string) types.Error {
	if name == "" {
		return types.ErrEmptyName()
	}
	if p == nil {
		return types.ErrNilPool()
	}
	p.Name = name
	return nil
}

func (p *Pool) SetAccount(account *Account) types.Error {
	if p == nil {
		return types.ErrNilPool()
	}
	if account == nil {
		return types.ErrEmptyAccount()
	}
	if err := account.ValidateBasic(); err != nil {
		return err
	}
	p.Account = account
	return nil
}
