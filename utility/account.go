package utility

// Internal business logic for `Accounts` & `Pools` (i.e. autonomous accounts owned by the protocol)
//
// Accounts are utility module structures that resemble currency holding vehicles; e.g. a bank account.
// Pools are autonomous accounts owned by the protocol; e.g. an account for a fee pool that gets distributed

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/pokterrors"
	"github.com/pokt-network/pocket/shared/utils"
)

// Accounts specific functionality

func (u *utilityContext) getAccountAmount(address []byte) (*big.Int, pokterrors.Error) {
	amountStr, err := u.store.GetAccountAmount(address, u.height)
	if err != nil {
		return nil, pokterrors.UtilityErrGetAccountAmount(err)
	}
	amount, err := utils.StringToBigInt(amountStr)
	if err != nil {
		return nil, pokterrors.UtilityErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *utilityContext) addAccountAmount(address []byte, amountToAdd *big.Int) pokterrors.Error {
	if err := u.store.AddAccountAmount(address, utils.BigIntToString(amountToAdd)); err != nil {
		return pokterrors.UtilityErrAddAccountAmount(err)
	}
	return nil
}

func (u *utilityContext) subtractAccountAmount(address []byte, amountToSubtract *big.Int) pokterrors.Error {
	if err := u.store.SubtractAccountAmount(address, utils.BigIntToString(amountToSubtract)); err != nil {
		return pokterrors.UtilityErrSetAccountAmount(err)
	}
	return nil
}

func (u *utilityContext) setAccountAmount(address []byte, amount *big.Int) pokterrors.Error {
	if err := u.store.SetAccountAmount(address, utils.BigIntToString(amount)); err != nil {
		return pokterrors.UtilityErrSetAccountAmount(err)
	}
	return nil
}

// Pools specific functionality

// IMPROVE: Pool function should accept the actual pool types rather than the `FriendlyName` string

func (u *utilityContext) insertPool(name string, amount *big.Int) pokterrors.Error {
	if err := u.store.InsertPool(name, utils.BigIntToString(amount)); err != nil {
		return pokterrors.UtilityErrSetPool(name, err)
	}
	return nil
}

func (u *utilityContext) getPoolAmount(name string) (*big.Int, pokterrors.Error) {
	amountStr, err := u.store.GetPoolAmount(name, u.height)
	if err != nil {
		return nil, pokterrors.UtilityErrGetPoolAmount(name, err)
	}
	amount, err := utils.StringToBigInt(amountStr)
	if err != nil {
		return nil, pokterrors.UtilityErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *utilityContext) addPoolAmount(name string, amountToAdd *big.Int) pokterrors.Error {
	if err := u.store.AddPoolAmount(name, utils.BigIntToString(amountToAdd)); err != nil {
		return pokterrors.UtilityErrAddPoolAmount(name, err)
	}
	return nil
}

func (u *utilityContext) subPoolAmount(name string, amountToSub *big.Int) pokterrors.Error {
	if err := u.store.SubtractPoolAmount(name, utils.BigIntToString(amountToSub)); err != nil {
		return pokterrors.UtilityErrSubPoolAmount(name, err)
	}
	return nil
}

func (u *utilityContext) setPoolAmount(name string, amount *big.Int) pokterrors.Error {
	if err := u.store.SetPoolAmount(name, utils.BigIntToString(amount)); err != nil {
		return pokterrors.UtilityErrSetPoolAmount(name, err)
	}
	return nil
}
