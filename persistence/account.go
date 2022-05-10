package persistence

import (
	"encoding/hex"
	"math/big"

	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	shared "github.com/pokt-network/pocket/shared/types"
)

// TODO(Andrew): Convert all queries to use 'height' in the interface for historical lookups
// TODO(Andrew): remove address from pool, use name only
// TODO(team): Consider converting all address params from bytes to string to avoid unnecessary encoding?

func (p PostgresContext) AddAccountAmount(address []byte, amount string) error {
	return p.operationAccountAmount(address, amount, func(s *big.Int, s1 *big.Int) error {
		s.Add(s, s1)
		return nil
	})
}

func (p PostgresContext) SubtractAccountAmount(address []byte, amount string) error {
	return p.operationAccountAmount(address, amount, func(s *big.Int, s1 *big.Int) error {
		s.Sub(s, s1)
		return nil
	})
}

func (p PostgresContext) GetAccountAmount(address []byte) (amount string, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	if err = conn.QueryRow(ctx, schema.GetAccountAmountQuery(hex.EncodeToString(address))).Scan(&amount); err != nil {
		return
	}
	return
}

func (p PostgresContext) SetAccountAmount(address []byte, amount string) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.NullifyAccountAmountQuery(hex.EncodeToString(address), height)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.InsertAccountAmountQuery(hex.EncodeToString(address), amount, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) GetPoolAmount(name string) (amount string, err error) {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return
	}
	if err = conn.QueryRow(ctx, schema.GetPoolAmountQuery(name)).Scan(&amount); err != nil {
		return
	}
	return
}

func (p PostgresContext) InsertPool(name string, address []byte, amount string) error { // TODO (Andrew) remove address param
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.NullifyPoolAmountQuery(name, height)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.InsertPoolAmountQuery(name, amount, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
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

func (p PostgresContext) SetPoolAmount(name string, amount string) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.NullifyPoolAmountQuery(name, height)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.InsertPoolAmountQuery(name, amount, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *PostgresContext) operationAccountAmount(address []byte, amount string, op func(*big.Int, *big.Int) error) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	originalAmount, err := p.GetAccountAmount(address)
	if err != nil {
		return err
	}
	originalAmountBig, err := shared.StringToBigInt(originalAmount)
	if err != nil {
		return err
	}
	amountBig, err := shared.StringToBigInt(amount)
	if err != nil {
		return err
	}
	if err := op(originalAmountBig, amountBig); err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, schema.NullifyAccountAmountQuery(hex.EncodeToString(address), height)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, schema.InsertAccountAmountQuery(hex.EncodeToString(address), shared.BigIntToString(originalAmountBig), height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *PostgresContext) operationPoolAmount(name string, amount string, op func(*big.Int, *big.Int) error) error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	originalAmount, err := p.GetPoolAmount(name)
	if err != nil {
		return err
	}
	originalAmountBig, err := shared.StringToBigInt(originalAmount)
	if err != nil {
		return err
	}
	amountBig, err := shared.StringToBigInt(amount)
	if err != nil {
		return err
	}
	if err := op(originalAmountBig, amountBig); err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.NullifyPoolAmountQuery(name, height)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.InsertPoolAmountQuery(name, shared.BigIntToString(originalAmountBig), height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
