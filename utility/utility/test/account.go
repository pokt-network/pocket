package test

import (
	"pocket/utility/shared/crypto"
	"pocket/utility/utility/types"
	"math/big"
)

func (x *Account) ValidateBasic() types.Error {
	if x == nil {
		return types.ErrEmptyAccount()
	}
	if x.Address == nil {
		return types.ErrEmptyAddress()
	}
	if len(x.Address) != crypto.AddressLen {
		return types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen())
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
	if len(x.Address) != crypto.AddressLen {
		return types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen())
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
