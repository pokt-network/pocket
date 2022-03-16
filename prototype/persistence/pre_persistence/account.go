package pre_persistence

import (
	"math/big"
	crypto2 "pocket/shared/crypto"
)

const (
	ServiceNodeStakePoolName = "SERVICE_NODE_STAKE_POOL"
	AppStakePoolName         = "APP_STAKE_POOL"
	ValidatorStakePoolName   = "VALIDATOR_STAKE_POOL"
	FishermanStakePoolName   = "FISHERMAN_STAKE_POOL"
	DAOPoolName              = "DAO_POOL"
	FeePoolName              = "FEE_POOL"
)

func (x *Account) ValidateBasic() Error {
	if x == nil {
		return ErrEmptyAccount()
	}
	if x.Address == nil {
		return ErrEmptyAddress()
	}
	if len(x.Address) != crypto2.AddressLen {
		return ErrInvalidAddressLen(crypto2.ErrInvalidAddressLen())
	}
	amount := big.Int{}
	if _, ok := amount.SetString(x.Amount, 10); !ok {
		return ErrInvalidAmount()
	}
	return nil
}

func (x *Account) SetAddress(address crypto2.Address) Error {
	if x == nil {
		return ErrEmptyAccount()
	}
	if len(x.Address) != crypto2.AddressLen {
		return ErrInvalidAddressLen(crypto2.ErrInvalidAddressLen())
	}
	x.Address = address
	return nil
}

func (x *Account) SetAmount(amount big.Int) Error {
	if x == nil {
		return ErrEmptyAccount()
	}
	x.Amount = amount.String()
	return nil
}

func NewPool(name string, account *Account) (*Pool, Error) {
	pool := &Pool{}
	if err := pool.SetName(name); err != nil {
		return nil, err
	}
	if err := pool.SetAccount(account); err != nil {
		return nil, err
	}
	return pool, nil
}

func (x *Pool) ValidateBasic() Error {
	if x == nil {
		return ErrNilPool()
	}
	if x.Name == "" {
		return ErrEmptyName()
	}
	return x.Account.ValidateBasic()
}

func (x *Pool) SetName(name string) Error {
	if name == "" {
		return ErrEmptyName()
	}
	if x == nil {
		return ErrNilPool()
	}
	x.Name = name
	return nil
}

func (x *Pool) SetAccount(account *Account) Error {
	if x == nil {
		return ErrNilPool()
	}
	if account == nil {
		return ErrEmptyAccount()
	}
	if err := account.ValidateBasic(); err != nil {
		return err
	}
	x.Account = account
	return nil
}
