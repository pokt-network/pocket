package utility

import (
	"math/big"
	types2 "pocket/utility/types"
)

func (u *UtilityContext) HandleMessageSend(message *types2.MessageSend) types2.Error {
	amount, err := StringToBigInt(message.Amount)
	if err != nil {
		return err
	}
	fromAccountAmount, err := u.GetAccountAmount(message.FromAddress)
	if err != nil {
		return err
	}
	fromAccountAmount.Sub(fromAccountAmount, amount)
	if fromAccountAmount.Sign() == -1 {
		return types2.ErrInsufficientAmountError()
	}
	if err := u.AddAccountAmount(message.ToAddress, amount); err != nil {
		return err
	}
	if err := u.SetAccount(message.FromAddress, fromAccountAmount); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) GetMessageSendSignerCandidates(msg *types2.MessageSend) (candidates [][]byte, err types2.Error) {
	return [][]byte{msg.FromAddress}, nil
}

func (u *UtilityContext) GetAccountAmount(address []byte) (*big.Int, types2.Error) {
	store := u.Store()
	amount, err := store.GetAccountAmount(address)
	if err != nil {
		return nil, types2.ErrGetAccountAmount(err)
	}
	return StringToBigInt(amount)
}

func (u *UtilityContext) AddAccountAmount(address []byte, amountToAdd *big.Int) types2.Error {
	store := u.Store()
	if err := store.AddAccountAmount(address, types2.BigIntToString(amountToAdd)); err != nil {
		return types2.ErrAddAccountAmount(err)
	}
	return nil
}

func (u *UtilityContext) AddAccountAmountString(address []byte, amountToAdd string) types2.Error {
	store := u.Store()
	if err := store.AddAccountAmount(address, amountToAdd); err != nil {
		return types2.ErrAddAccountAmount(err)
	}
	return nil
}

func (u *UtilityContext) AddPoolAmount(name string, amountToAdd *big.Int) types2.Error {
	store := u.Store()
	if err := store.AddPoolAmount(name, types2.BigIntToString(amountToAdd)); err != nil {
		return types2.ErrAddPoolAmount(name, err)
	}
	return nil
}

func (u *UtilityContext) SubPoolAmount(name string, amountToSub string) types2.Error {
	store := u.Store()
	if err := store.SubtractPoolAmount(name, amountToSub); err != nil {
		return types2.ErrSubPoolAmount(name, err)
	}
	return nil
}

func (u *UtilityContext) GetPoolAmount(name string) (amount *big.Int, err types2.Error) {
	store := u.Store()
	tokens, er := store.GetPoolAmount(name)
	if er != nil {
		return nil, types2.ErrGetPoolAmount(name, er)
	}
	amount, err = StringToBigInt(tokens)
	if err != nil {
		return nil, err
	}
	return
}

func (u *UtilityContext) InsertPool(name string, address []byte, amount string) (err types2.Error) {
	store := u.Store()
	if err := store.InsertPool(name, address, amount); err != nil {
		return types2.ErrSetPool(name, err)
	}
	return
}

func (u *UtilityContext) SetPoolAmount(name string, amount *big.Int) (err types2.Error) {
	store := u.Store()
	if err := store.SetPoolAmount(name, BigIntToString(amount)); err != nil {
		return types2.ErrSetPoolAmount(name, err)
	}
	return
}

func (u *UtilityContext) SetAccountWithAmountString(address []byte, amount string) types2.Error {
	store := u.Store()
	if err := store.SetAccount(address, amount); err != nil {
		return types2.ErrSetAccount(err)
	}
	return nil
}

func (u *UtilityContext) SetAccount(address []byte, amount *big.Int) types2.Error {
	store := u.Store()
	if err := store.SetAccount(address, types2.BigIntToString(amount)); err != nil {
		return types2.ErrSetAccount(err)
	}
	return nil
}

func (u *UtilityContext) SubtractAccountAmount(address []byte, amountToSubtract *big.Int) types2.Error {
	store := u.Store()
	if err := store.SubtractAccountAmount(address, BigIntToString(amountToSubtract)); err != nil {
		return types2.ErrSetAccount(err)
	}
	return nil
}
