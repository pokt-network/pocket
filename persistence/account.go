package persistence

import (
	"encoding/hex"
	"math/big"

	"github.com/pokt-network/pocket/persistence/types"

	"github.com/jackc/pgx/v4"
)

// TODO(https://github.com/pokt-network/pocket/issues/102): Generalize Pool and Account operations.

const (
	defaultAccountAmountStr string = "0"
)

// --- Account Functions ---

//
func (p PostgresContext) GetAccountAmount(address []byte, height int64) (amount string, err error) {
	return p.getAccountAmountStr(hex.EncodeToString(address), height)
}

func (p PostgresContext) getAccountAmountStr(address string, height int64) (amount string, err error) {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return
	}

	amount = defaultAccountAmountStr
	if err = txn.QueryRow(ctx, types.GetAccountAmountQuery(address, height)).Scan(&amount); err != pgx.ErrNoRows {
		return
	}

	return amount, nil
}

func (p PostgresContext) AddAccountAmount(address []byte, amount string) error {
	return p.operationAccountAmount(address, amount, func(orig *big.Int, delta *big.Int) error {
		orig.Add(orig, delta)
		return nil
	})
}

func (p PostgresContext) SubtractAccountAmount(address []byte, amount string) error {
	return p.operationAccountAmount(address, amount, func(orig *big.Int, delta *big.Int) error {
		orig.Sub(orig, delta)
		return nil
	})
}

// DISCUSS(team): If we are okay with `GetAccountAmount` return 0 as a default, this function can leverage
//                `operationAccountAmount` with `*orig = *delta` and make everything much simpler.
func (p PostgresContext) SetAccountAmount(address []byte, amount string) error {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	// DISCUSS(team): Do we want to panic if `amount < 0` here?
	if _, err = txn.Exec(ctx, types.InsertAccountAmountQuery(hex.EncodeToString(address), amount, height)); err != nil {
		return err
	}
	return nil
}

func (p *PostgresContext) operationAccountAmount(address []byte, deltaAmount string, op func(*big.Int, *big.Int) error) error {
	return p.operationPoolOrAccAmount(hex.EncodeToString(address), deltaAmount, op, p.getAccountAmountStr, types.InsertAccountAmountQuery)
}

// --- Pool Functions ---

// TODO(andrew): remove address param
func (p PostgresContext) InsertPool(name string, address []byte, amount string) error {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	if _, err = txn.Exec(ctx, types.InsertPoolAmountQuery(name, amount, height)); err != nil {
		return err
	}
	return nil
}

func (p PostgresContext) GetPoolAmount(name string, height int64) (amount string, err error) {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return
	}

	amount = defaultAccountAmountStr
	if err = txn.QueryRow(ctx, types.GetPoolAmountQuery(name, height)).Scan(&amount); err != pgx.ErrNoRows {
		return
	}

	return amount, nil
}

func (p PostgresContext) AddPoolAmount(name string, amount string) error {
	return p.operationPoolAmount(name, amount, func(s *big.Int, s1 *big.Int) error {
		s.Add(s, s1)
		return nil
	})
}

func (p PostgresContext) SubtractPoolAmount(name string, amount string) error {
	return p.operationPoolAmount(name, amount, func(s *big.Int, s1 *big.Int) error {
		s.Sub(s, s1)
		return nil
	})
}

// DISCUSS(team): If we are okay with `GetPoolAmount` return 0 as a default, this function can leverage
//                `operationPoolAmount` with `*orig = *delta` and make everything much simpler.
// DISCUSS(team): Do we have a use-case for this function?
func (p PostgresContext) SetPoolAmount(name string, amount string) error {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	if _, err = txn.Exec(ctx, types.InsertPoolAmountQuery(name, amount, height)); err != nil {
		return err
	}
	return nil
}

func (p *PostgresContext) operationPoolAmount(name string, amount string, op func(*big.Int, *big.Int) error) error {
	return p.operationPoolOrAccAmount(name, amount, op, p.GetPoolAmount, types.InsertPoolAmountQuery)
}

func (p *PostgresContext) operationPoolOrAccAmount(name, amount string,
	op func(*big.Int, *big.Int) error,
	getAmount func(string, int64) (string, error),
	insert func(name, amount string, height int64) string) error {
	ctx, txn, err := p.GetCtxAndTxn()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	originalAmount, err := getAmount(name, height)
	if err != nil {
		return err
	}
	originalAmountBig, err := types.StringToBigInt(originalAmount)
	if err != nil {
		return err
	}
	amountBig, err := types.StringToBigInt(amount)
	if err != nil {
		return err
	}
	if err := op(originalAmountBig, amountBig); err != nil {
		return err
	}
	if _, err = txn.Exec(ctx, insert(name, types.BigIntToString(originalAmountBig), height)); err != nil {
		return err
	}
	return nil
}
