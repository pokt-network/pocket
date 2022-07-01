package pre_persistence

import (
	"bytes"

	"github.com/pokt-network/pocket/shared/types"

	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"

	"math/big"

	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
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
	codec := types.GetCodec()
	p := typesGenesis.Pool{}
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
	s, err := types.StringToBigInt(p.Account.Amount)
	if err != nil {
		return err
	}
	s2, err := types.StringToBigInt(amount)
	if err != nil {
		return err
	}
	if err := op(s, s2); err != nil {
		return err
	}
	p.Account.Amount = types.BigIntToString(s)
	bz, err := codec.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Put(key, bz)
}

func (m *PrePersistenceContext) InsertPool(name string, address []byte, amount string) error {
	codec := types.GetCodec()
	p := typesGenesis.Pool{
		Name: name,
		Account: &typesGenesis.Account{
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
	codec := types.GetCodec()
	p := typesGenesis.Pool{}
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

func (m *PrePersistenceContext) GetPoolAmount(name string, height int64) (amount string, err error) {
	codec := types.GetCodec()
	p := typesGenesis.Pool{}
	db := m.Store()
	key := append(PoolPrefixKey, []byte(name)...)
	val, err := db.Get(key)
	if err != nil {
		return types.EmptyString, err
	}
	err = codec.Unmarshal(val, &p)
	if err != nil {
		return types.EmptyString, err
	}
	return p.Account.Amount, nil
}

func (m *PrePersistenceContext) GetAllPools(height int64) (pools []*typesGenesis.Pool, err error) {
	codec := types.GetCodec()
	pools = make([]*typesGenesis.Pool, 0)
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
		p := typesGenesis.Pool{}
		if err := codec.Unmarshal(bz, &p); err != nil {
			return nil, err
		}
		pools = append(pools, &p)
	}
	return
}

func (m *PrePersistenceContext) operationAccountAmount(address []byte, amount string, op func(a, b *big.Int) error) error {
	codec := types.GetCodec()
	a := typesGenesis.Account{}
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
	s, err := types.StringToBigInt(a.Amount)
	if err != nil {
		return err
	}
	s2, err := types.StringToBigInt(amount)
	if err != nil {
		return err
	}
	if err := op(s, s2); err != nil {
		return err
	}
	a.Amount = types.BigIntToString(s)
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

func (m *PrePersistenceContext) GetAccountAmount(address []byte, height int64) (string, error) {
	codec := types.GetCodec()
	account := typesGenesis.Account{}
	db := m.Store()
	key := append(AccountPrefixKey, address...)
	val, err := db.Get(key)
	if err != nil {
		return types.EmptyString, err
	}
	err = codec.Unmarshal(val, &account)
	if err != nil {
		return types.EmptyString, err
	}
	return account.Amount, nil
}

func (m *PrePersistenceContext) SetAccountAmount(address []byte, amount string) error {
	codec := types.GetCodec()
	account := typesGenesis.Account{
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

func (m *PrePersistenceContext) GetAllAccounts(height int64) (accs []*typesGenesis.Account, err error) {
	codec := types.GetCodec()
	accs = make([]*typesGenesis.Account, 0)
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
		acc := typesGenesis.Account{}
		if err := codec.Unmarshal(bz, &acc); err != nil {
			return nil, err
		}
		accs = append(accs, &acc)
	}
	return
}
