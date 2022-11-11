package persistence

import (
	"encoding/hex"
	"math/big"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/converters"

	"github.com/jackc/pgx/v4"
)

// TODO(https://github.com/pokt-network/pocket/issues/102): Generalize Pool and Account operations.

const (
	defaultAccountAmountStr string = "0"
)

// --- Account Functions ---

func (p PostgresContext) GetAccountAmount(address []byte, height int64) (amount string, err error) {
	return p.getAccountAmountStr(hex.EncodeToString(address), height)
}

func (p PostgresContext) getAccountAmountStr(address string, height int64) (amount string, err error) {
	ctx, tx, err := p.GetCtxAndTx()
	if err != nil {
		return
	}

	amount = defaultAccountAmountStr
	if err = tx.QueryRow(ctx, types.GetAccountAmountQuery(address, height)).Scan(&amount); err != pgx.ErrNoRows {
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
	ctx, tx, err := p.GetCtxAndTx()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	// DISCUSS(team): Do we want to panic if `amount < 0` here?
	if _, err = tx.Exec(ctx, types.InsertAccountAmountQuery(hex.EncodeToString(address), amount, height)); err != nil {
		return err
	}
	return nil
}

func (p PostgresContext) GetAccountsUpdated(height int64) (accounts []*types.Account, err error) {
	return p.getPoolOrAccUpdatedInternal(types.GetAccountsUpdatedAtHeightQuery(height))
}

func (p *PostgresContext) operationAccountAmount(address []byte, deltaAmount string, op func(*big.Int, *big.Int) error) error {
	return p.operationPoolOrAccAmount(hex.EncodeToString(address), deltaAmount, op, p.getAccountAmountStr, types.InsertAccountAmountQuery)
}

// --- Pool Functions ---

// TODO(andrew): remove address param
func (p PostgresContext) InsertPool(name string, address []byte, amount string) error {
	ctx, tx, err := p.GetCtxAndTx()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, types.InsertPoolAmountQuery(name, amount, height)); err != nil {
		return err
	}
	return nil
}

func (p PostgresContext) GetPoolAmount(name string, height int64) (amount string, err error) {
	ctx, tx, err := p.GetCtxAndTx()
	if err != nil {
		return
	}

	amount = defaultAccountAmountStr
	if err = tx.QueryRow(ctx, types.GetPoolAmountQuery(name, height)).Scan(&amount); err != pgx.ErrNoRows {
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
//
//	`operationPoolAmount` with `*orig = *delta` and make everything much simpler.
//
// DISCUSS(team): Do we have a use-case for this function?
func (p PostgresContext) SetPoolAmount(name string, amount string) error {
	ctx, tx, err := p.GetCtxAndTx()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, types.InsertPoolAmountQuery(name, amount, height)); err != nil {
		return err
	}
	return nil
}

func (p *PostgresContext) operationPoolAmount(name string, amount string, op func(*big.Int, *big.Int) error) error {
	return p.operationPoolOrAccAmount(name, amount, op, p.GetPoolAmount, types.InsertPoolAmountQuery)
}

func (p PostgresContext) GetPoolsUpdated(height int64) ([]*types.Account, error) {
	return p.getPoolOrAccUpdatedInternal(types.GetPoolsUpdatedAtHeightQuery(height))
}

// Joint Pool & Account Helpers

// Helper for shared logic between `getPoolsUpdated` and `getAccountsUpdated` while keeping an explicit
// external interface.
func (p *PostgresContext) getPoolOrAccUpdatedInternal(query string) (accounts []*types.Account, err error) {
	ctx, tx, err := p.GetCtxAndTx()
	if err != nil {
		return
	}

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		account := new(types.Account)
		if err = rows.Scan(&account.Address, &account.Amount); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return
}

func (p *PostgresContext) operationPoolOrAccAmount(
	name, amount string,
	op func(*big.Int, *big.Int) error,
	getAmount func(string, int64) (string, error),
	insert func(name, amount string, height int64) string,
) error {
	ctx, tx, err := p.GetCtxAndTx()
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
	originalAmountBig, err := converters.StringToBigInt(originalAmount)
	if err != nil {
		return err
	}
	amountBig, err := converters.StringToBigInt(amount)
	if err != nil {
		return err
	}
	if err := op(originalAmountBig, amountBig); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, insert(name, converters.BigIntToString(originalAmountBig), height)); err != nil {
		return err
	}
	return nil
}
