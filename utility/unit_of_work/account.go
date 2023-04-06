package unit_of_work

// Internal business logic for `Accounts` & `Pools` (i.e. autonomous accounts owned by the protocol)
//
// Accounts are utility module structures that resemble currency holding vehicles; e.g. a bank account.
// Pools are autonomous accounts owned by the protocol; e.g. an account for a fee pool that gets distributed

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/utils"
	"github.com/pokt-network/pocket/utility/types"
)

// Accounts specific functionality

func (u *baseUtilityUnitOfWork) getAccountAmount(address []byte) (*big.Int, types.Error) {
	amountStr, err := u.persistenceReadContext.GetAccountAmount(address, u.height)
	if err != nil {
		return nil, types.ErrGetAccountAmount(err)
	}
	amount, err := utils.StringToBigInt(amountStr)
	if err != nil {
		return nil, types.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *baseUtilityUnitOfWork) addAccountAmount(address []byte, amountToAdd *big.Int) types.Error {
	if err := u.persistenceRWContext.AddAccountAmount(address, utils.BigIntToString(amountToAdd)); err != nil {
		return types.ErrAddAccountAmount(err)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) subtractAccountAmount(address []byte, amountToSubtract *big.Int) types.Error {
	if err := u.persistenceRWContext.SubtractAccountAmount(address, utils.BigIntToString(amountToSubtract)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) setAccountAmount(address []byte, amount *big.Int) types.Error {
	if err := u.persistenceRWContext.SetAccountAmount(address, utils.BigIntToString(amount)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}

// Pools specific functionality

func (u *baseUtilityUnitOfWork) insertPool(address []byte, amount *big.Int) types.Error {
	if err := u.persistenceRWContext.InsertPool(address, utils.BigIntToString(amount)); err != nil {
		return types.ErrSetPool(address, err)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) getPoolAmount(address []byte) (*big.Int, types.Error) {
	amountStr, err := u.persistenceReadContext.GetPoolAmount(address, u.height)
	if err != nil {
		return nil, types.ErrGetPoolAmount(address, err)
	}
	amount, err := utils.StringToBigInt(amountStr)
	if err != nil {
		return nil, types.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *baseUtilityUnitOfWork) addPoolAmount(address []byte, amountToAdd *big.Int) types.Error {
	if err := u.persistenceRWContext.AddPoolAmount(address, utils.BigIntToString(amountToAdd)); err != nil {
		return types.ErrAddPoolAmount(address, err)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) subPoolAmount(address []byte, amountToSub *big.Int) types.Error {
	if err := u.persistenceRWContext.SubtractPoolAmount(address, utils.BigIntToString(amountToSub)); err != nil {
		return types.ErrSubPoolAmount(address, err)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) setPoolAmount(address []byte, amount *big.Int) types.Error {
	if err := u.persistenceRWContext.SetPoolAmount(address, utils.BigIntToString(amount)); err != nil {
		return types.ErrSetPoolAmount(address, err)
	}
	return nil
}
