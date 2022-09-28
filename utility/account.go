package utility

import (
	"github.com/pokt-network/pocket/utility/types"
	"math/big"
)

// 'Accounts' are structures in the utility module that closely resemble currency holding vehicles: like a bank account.
//  Accounts enable the 'ownership' or 'custody' over uPOKT tokens. These structures are fundamental to enabling
//  the utility economy.

func (u *UtilityContext) GetAccountAmount(address []byte) (*big.Int, types.Error) {
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return nil, er
	}
	amount, err := store.GetAccountAmount(address, height)
	if err != nil {
		return nil, types.ErrGetAccountAmount(err)
	}
	return types.StringToBigInt(amount)
}

func (u *UtilityContext) AddAccountAmount(address []byte, amountToAdd *big.Int) types.Error {
	store := u.Store()
	if err := store.AddAccountAmount(address, types.BigIntToString(amountToAdd)); err != nil {
		return types.ErrAddAccountAmount(err)
	}
	return nil
}

func (u *UtilityContext) AddAccountAmountString(address []byte, amountToAdd string) types.Error {
	store := u.Store()
	if err := store.AddAccountAmount(address, amountToAdd); err != nil {
		return types.ErrAddAccountAmount(err)
	}
	return nil
}

func (u *UtilityContext) AddPoolAmount(name string, amountToAdd *big.Int) types.Error {
	store := u.Store()
	if err := store.AddPoolAmount(name, types.BigIntToString(amountToAdd)); err != nil {
		return types.ErrAddPoolAmount(name, err)
	}
	return nil
}

func (u *UtilityContext) SubPoolAmount(name string, amountToSub string) types.Error {
	store := u.Store()
	if err := store.SubtractPoolAmount(name, amountToSub); err != nil {
		return types.ErrSubPoolAmount(name, err)
	}
	return nil
}

func (u *UtilityContext) GetPoolAmount(name string) (*big.Int, types.Error) {
	store, height, err := u.GetStoreAndHeight()
	if err != nil {
		return nil, err
	}
	tokens, er := store.GetPoolAmount(name, height)
	if er != nil {
		return nil, types.ErrGetPoolAmount(name, er)
	}
	amount, err := types.StringToBigInt(tokens)
	if err != nil {
		return nil, err
	}
	return amount, nil
}

func (u *UtilityContext) InsertPool(name string, address []byte, amount string) types.Error {
	store := u.Store()
	if err := store.InsertPool(name, address, amount); err != nil {
		return types.ErrSetPool(name, err)
	}
	return nil
}

func (u *UtilityContext) SetPoolAmount(name string, amount *big.Int) types.Error {
	store := u.Store()
	if err := store.SetPoolAmount(name, types.BigIntToString(amount)); err != nil {
		return types.ErrSetPoolAmount(name, err)
	}
	return nil
}

func (u *UtilityContext) SetAccountWithAmountString(address []byte, amount string) types.Error {
	store := u.Store()
	if err := store.SetAccountAmount(address, amount); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}

func (u *UtilityContext) SetAccountAmount(address []byte, amount *big.Int) types.Error {
	store := u.Store()
	if err := store.SetAccountAmount(address, types.BigIntToString(amount)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}

func (u *UtilityContext) SubtractAccountAmount(address []byte, amountToSubtract *big.Int) types.Error {
	store := u.Store()
	if err := store.SubtractAccountAmount(address, types.BigIntToString(amountToSubtract)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}
