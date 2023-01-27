package utility

import (
	"math/big"

	"github.com/pokt-network/pocket/utility/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// 'Accounts' are structures in the utility module that closely resemble currency holding vehicles: like a bank account.
//  Accounts enable the 'ownership' or 'custody' over uPOKT tokens. These structures are fundamental to enabling
//  the utility economy.

func (u *utilityContext) getAccountAmount(address []byte) (*big.Int, types.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}
	amount, err := store.GetAccountAmount(address, height)
	if err != nil {
		return nil, typesUtil.ErrGetAccountAmount(err)
	}
	return typesUtil.StringToBigInt(amount)
}

func (u *utilityContext) addAccountAmount(address []byte, amountToAdd *big.Int) types.Error {
	store := u.Store()
	if err := store.AddAccountAmount(address, types.BigIntToString(amountToAdd)); err != nil {
		return types.ErrAddAccountAmount(err)
	}
	return nil
}

func (u *utilityContext) addAccountAmountString(address []byte, amountToAdd string) types.Error {
	store := u.Store()
	if err := store.AddAccountAmount(address, amountToAdd); err != nil {
		return types.ErrAddAccountAmount(err)
	}
	return nil
}

func (u *utilityContext) addPoolAmount(name string, amountToAdd *big.Int) types.Error {
	store := u.Store()
	if err := store.AddPoolAmount(name, types.BigIntToString(amountToAdd)); err != nil {
		return types.ErrAddPoolAmount(name, err)
	}
	return nil
}

func (u *utilityContext) SubPoolAmount(name string, amountToSub string) types.Error {
	store := u.Store()
	if err := store.SubtractPoolAmount(name, amountToSub); err != nil {
		return types.ErrSubPoolAmount(name, err)
	}
	return nil
}

func (u *utilityContext) GetPoolAmount(name string) (*big.Int, types.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}
	tokens, err := store.GetPoolAmount(name, height)
	if err != nil {
		return nil, types.ErrGetPoolAmount(name, err)
	}
	return types.StringToBigInt(tokens)
}

func (u *utilityContext) InsertPool(name string, address []byte, amount string) types.Error {
	store := u.Store()
	if err := store.InsertPool(name, amount); err != nil {
		return types.ErrSetPool(name, err)
	}
	return nil
}

func (u *utilityContext) SetPoolAmount(name string, amount *big.Int) types.Error {
	store := u.Store()
	if err := store.SetPoolAmount(name, types.BigIntToString(amount)); err != nil {
		return types.ErrSetPoolAmount(name, err)
	}
	return nil
}

func (u *utilityContext) SetAccountWithAmountString(address []byte, amount string) types.Error {
	store := u.Store()
	if err := store.SetAccountAmount(address, amount); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}

func (u *utilityContext) SetAccountAmount(address []byte, amount *big.Int) types.Error {
	store := u.Store()
	if err := store.SetAccountAmount(address, types.BigIntToString(amount)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}

func (u *utilityContext) SubtractAccountAmount(address []byte, amountToSubtract *big.Int) types.Error {
	store := u.Store()
	if err := store.SubtractAccountAmount(address, types.BigIntToString(amountToSubtract)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}
