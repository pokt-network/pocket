package utility

// Internal business logic for `Accounts` & `Pools` (i.e. autonomous accounts owned by the protocol)
//
// Accounts are utility module structures that resemble currency holding vehicles; e.g. a bank account.
// Pools are autonomous accounts owned by the protocol; e.g. an account for a fee pool that gets distributed

import (
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
)

// Accounts specific functionality

func (u *utilityContext) getAccountAmount(address []byte) (*big.Int, coreTypes.Error) {
	amountStr, err := u.store.GetAccountAmount(address, u.height)
	if err != nil {
		return nil, coreTypes.ErrGetAccountAmount(err)
	}
	amount, err := utils.StringToBigInt(amountStr)
	if err != nil {
		return nil, coreTypes.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *utilityContext) addAccountAmount(address []byte, amountToAdd *big.Int) coreTypes.Error {
	if err := u.store.AddAccountAmount(address, utils.BigIntToString(amountToAdd)); err != nil {
		return coreTypes.ErrAddAccountAmount(err)
	}
	return nil
}

func (u *utilityContext) subtractAccountAmount(address []byte, amountToSubtract *big.Int) coreTypes.Error {
	if err := u.store.SubtractAccountAmount(address, utils.BigIntToString(amountToSubtract)); err != nil {
		return coreTypes.ErrSetAccountAmount(err)
	}
	return nil
}

func (u *utilityContext) setAccountAmount(address []byte, amount *big.Int) coreTypes.Error {
	if err := u.store.SetAccountAmount(address, utils.BigIntToString(amount)); err != nil {
		return coreTypes.ErrSetAccountAmount(err)
	}
	return nil
}

// Pools specific functionality

// IMPROVE: Pool function should accept the actual pool types rather than the `FriendlyName` string

func (u *utilityContext) insertPool(name string, amount *big.Int) coreTypes.Error {
	if err := u.store.InsertPool(name, utils.BigIntToString(amount)); err != nil {
		return coreTypes.ErrSetPool(name, err)
	}
	return nil
}

func (u *utilityContext) getPoolAmount(name string) (*big.Int, coreTypes.Error) {
	amountStr, err := u.store.GetPoolAmount(name, u.height)
	if err != nil {
		return nil, coreTypes.ErrGetPoolAmount(name, err)
	}
	amount, err := utils.StringToBigInt(amountStr)
	if err != nil {
		return nil, coreTypes.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *utilityContext) addPoolAmount(name string, amountToAdd *big.Int) coreTypes.Error {
	if err := u.store.AddPoolAmount(name, utils.BigIntToString(amountToAdd)); err != nil {
		return coreTypes.ErrAddPoolAmount(name, err)
	}
	return nil
}

func (u *utilityContext) subPoolAmount(name string, amountToSub *big.Int) coreTypes.Error {
	if err := u.store.SubtractPoolAmount(name, utils.BigIntToString(amountToSub)); err != nil {
		return coreTypes.ErrSubPoolAmount(name, err)
	}
	return nil
}

func (u *utilityContext) setPoolAmount(name string, amount *big.Int) coreTypes.Error {
	if err := u.store.SetPoolAmount(name, utils.BigIntToString(amount)); err != nil {
		return coreTypes.ErrSetPoolAmount(name, err)
	}
	return nil
}
