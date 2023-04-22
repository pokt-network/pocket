package kvstore

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresKV implements the KVStore interface.
type PostgresKV struct {
	Pool *pgxpool.Pool
}

// Get returns the value at a given key or an error.
func (p *PostgresKV) Get(key []byte) ([]byte, error) {
	ctx := context.TODO()
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Conn().Close(ctx)

	fmt.Printf("kv store connected to pool %+v", conn)

	return nil, fmt.Errorf("not implemented") // TODO: Implement
}

// Set ...
func (p *PostgresKV) Set(key []byte, value []byte) error {
	ctx := context.TODO()
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Conn().Close(ctx)

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	res, err := tx.Exec(ctx, "INSERT INTO transactions values ($1, $2)", key, value)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	fmt.Printf("inserted %d", res.RowsAffected())
	fmt.Printf("res#String: %v", res.String)

	return nil
}

// Delete ...
func (p *PostgresKV) Delete(key []byte) error {
	panic("not implemented") // TODO: Implement
}

// Lifecycle methods

// Stop ...
func (p *PostgresKV) Stop() error {
	panic("not implemented") // TODO: Implement
}

// Accessors

// GetAll ...
func (p *PostgresKV) GetAll(prefixKey []byte, descending bool) (keys [][]byte, values [][]byte, err error) {
	panic("not implemented") // TODO: Implement
}

// Exists ...
func (p *PostgresKV) Exists(key []byte) (bool, error) {
	panic("not implemented") // TODO: Implement
}

// ClearAll ...
func (p *PostgresKV) ClearAll() error {
	panic("not implemented") // TODO: Implement
}
