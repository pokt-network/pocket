package utility

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/converters"
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
	amountStr, err := store.GetAccountAmount(address, height)
	if err != nil {
		return nil, typesUtil.ErrGetAccountAmount(err)
	}
	amount, err := converters.StringToBigInt(amountStr)
	if err != nil {
		return nil, typesUtil.ErrStringToBigInt()
	}
	return amount, nil
}

func (u *utilityContext) addAccountAmount(address []byte, amountToAdd *big.Int) types.Error {
	store := u.Store()
	if err := store.AddAccountAmount(address, converters.BigIntToString(amountToAdd)); err != nil {
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
	if err := store.AddPoolAmount(name, converters.BigIntToString(amountToAdd)); err != nil {
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
	amount, err := converters.StringToBigInt(tokens)
	if err != nil {
		return nil, types.ErrStringToBigInt()
	}
	return amount, nil
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
	if err := store.SetPoolAmount(name, converters.BigIntToString(amount)); err != nil {
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
	if err := store.SetAccountAmount(address, converters.BigIntToString(amount)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}

func (u *utilityContext) SubtractAccountAmount(address []byte, amountToSubtract *big.Int) types.Error {
	store := u.Store()
	if err := store.SubtractAccountAmount(address, converters.BigIntToString(amountToSubtract)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}
