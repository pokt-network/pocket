package utility

// Internal business logic for `Accounts` & `Pools` (i.e. autonomous accounts owned by the protocol)
//
// Accounts are utility module structures that resemble currency holding vehicles; e.g. a bank account.
// Pools are autonomous accounts owned by the protocol; e.g. an account for a fee pool that gets distributed

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/utils"
	"github.com/pokt-network/pocket/utility/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// Accounts specific functionality

func (u *utilityContext) getAccountAmount(address []byte) (*big.Int, types.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}
	amountStr, err := store.GetAccountAmount(address, height)
	if err != nil {
		return nil, typesUtil.ErrGetAccountAmount(err)
	}
	amount, err := utils.StringToBigInt(amountStr)
	if err != nil {
		return nil, typesUtil.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *utilityContext) addAccountAmount(address []byte, amountToAdd *big.Int) types.Error {
	if err := u.Store().AddAccountAmount(address, utils.BigIntToString(amountToAdd)); err != nil {
		return types.ErrAddAccountAmount(err)
	}
	return nil
}

func (u *utilityContext) subtractAccountAmount(address []byte, amountToSubtract *big.Int) types.Error {
	if err := u.Store().SubtractAccountAmount(address, utils.BigIntToString(amountToSubtract)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}

func (u *utilityContext) setAccountAmount(address []byte, amount *big.Int) types.Error {
	if err := u.Store().SetAccountAmount(address, utils.BigIntToString(amount)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}

// Pools specific functionality

func (u *utilityContext) insertPool(name string, amount *big.Int) types.Error {
	if err := u.Store().InsertPool(name, utils.BigIntToString(amount)); err != nil {
		return types.ErrSetPool(name, err)
	}
	return nil
}

func (u *utilityContext) getPoolAmount(name string) (*big.Int, types.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}
	amountStr, err := store.GetPoolAmount(name, height)
	if err != nil {
		return nil, types.ErrGetPoolAmount(name, err)
	}
	amount, err := utils.StringToBigInt(amountStr)
	if err != nil {
		return nil, types.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *utilityContext) addPoolAmount(name string, amountToAdd *big.Int) types.Error {
	if err := u.Store().AddPoolAmount(name, utils.BigIntToString(amountToAdd)); err != nil {
		return types.ErrAddPoolAmount(name, err)
	}
	return nil
}

func (u *utilityContext) subPoolAmount(name string, amountToSub *big.Int) types.Error {
	if err := u.Store().SubtractPoolAmount(name, utils.BigIntToString(amountToSub)); err != nil {
		return types.ErrSubPoolAmount(name, err)
	}
	return nil
}

func (u *utilityContext) setPoolAmount(name string, amount *big.Int) types.Error {
	if err := u.Store().SetPoolAmount(name, utils.BigIntToString(amount)); err != nil {
		return types.ErrSetPoolAmount(name, err)
	}
	return nil
}
