package genesis

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
)

func (a *Account) ValidateBasic() types.Error {
	if a == nil {
		return types.ErrEmptyAccount()
	}
	if a.Address == nil {
		return types.ErrEmptyAddress()
	}
	if len(a.Address) != crypto.AddressLen {
		return types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen(crypto.AddressLen))
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
		return types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen(crypto.AddressLen))
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
