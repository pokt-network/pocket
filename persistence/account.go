package persistence

import (
	"encoding/hex"
	"math/big"

	"github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// TODO(https://github.com/pokt-network/pocket/issues/102): Generalize Pool and Account operations.

// --- Account Functions ---

func (p PostgresContext) GetAccountAmount(address []byte, height int64) (amount string, err error) {
	return p.getAccountAmount(types.Account, hex.EncodeToString(address), height)
}

func (p PostgresContext) AddAccountAmount(address []byte, amount string) error {
	return p.operationAccountAmount(types.Account, hex.EncodeToString(address), amount, func(orig *big.Int, delta *big.Int) error {
		orig.Add(orig, delta)
		return nil
	})
}

func (p PostgresContext) SubtractAccountAmount(address []byte, amount string) error {
	return p.operationAccountAmount(types.Account, hex.EncodeToString(address), amount, func(orig *big.Int, delta *big.Int) error {
		orig.Sub(orig, delta)
		return nil
	})
}

// DISCUSS(team): If we are okay with `GetAccountAmount` return 0 as a default, this function can leverage
// `operationAccountAmount` with `*orig = *delta` and make everything much simpler.
func (p PostgresContext) SetAccountAmount(address []byte, amount string) error {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	// DISCUSS(team): Do we want to panic if `amount < 0` here?
	if _, err = tx.Exec(ctx, types.Account.InsertAccountQuery(hex.EncodeToString(address), amount, height)); err != nil {
		return err
	}
	return nil
}

func (p PostgresContext) GetAccountsUpdated(height int64) (accounts []*coreTypes.Account, err error) {
	return p.getAccountsUpdated(types.Account, height)
}

// --- Pool Functions ---

func (p PostgresContext) InsertPool(name string, amount string) error {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, types.Pool.InsertAccountQuery(name, amount, height)); err != nil {
		return err
	}
	return nil
}

func (p PostgresContext) GetPoolAmount(name string, height int64) (amount string, err error) {
	return p.getAccountAmount(types.Pool, name, height)
}

func (p PostgresContext) AddPoolAmount(name string, amount string) error {
	return p.operationAccountAmount(types.Pool, name, amount, func(orig *big.Int, delta *big.Int) error {
		orig.Add(orig, delta)
		return nil
	})
}

func (p PostgresContext) SubtractPoolAmount(name string, amount string) error {
	return p.operationAccountAmount(types.Pool, name, amount, func(orig *big.Int, delta *big.Int) error {
		orig.Sub(orig, delta)
		return nil
	})
}

// DISCUSS(team): If we are okay with `GetPoolAmount` return 0 as a default, this function can leverage
//
//	`operationAccountAmount` with `*orig = *delta` and make everything much simpler.
//
// DISCUSS(team): Do we have a use-case for this function?
func (p PostgresContext) SetPoolAmount(name string, amount string) error {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, types.Pool.InsertAccountQuery(name, amount, height)); err != nil {
		return err
	}
	return nil
}

func (p PostgresContext) GetPoolsUpdated(height int64) ([]*coreTypes.Account, error) {
	return p.getAccountsUpdated(types.Pool, height)
}
