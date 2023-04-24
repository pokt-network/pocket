package kvstore

import (
	"context"
	"database/sql"
	"errors"
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

	row := conn.QueryRow(ctx, "SELECT FROM transactions WHERE key=$1", key)
	var val []byte
	if err := row.Scan(&val); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrKVStoreNotExists
		}
		return nil, err
	}
	return val, nil
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
	fmt.Printf("res#String: %s", res.String())

	return nil
}

// Delete ...
func (p *PostgresKV) Delete(key []byte) error {
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

	res, err := tx.Exec(ctx, "DELETE FROM transactions ($1);", key)
	if err != nil {
		return err
	}

	fmt.Printf("res.Delete(): %v\n", res.Delete())

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// Lifecycle methods

// Stop ...
func (p *PostgresKV) Stop() error {
	return nil
}

// Accessors

// GetAll gets all of the keys that fit the prefix
func (p *PostgresKV) GetAll(prefixKey []byte, descending bool) (keys [][]byte, values [][]byte, err error) {
	ctx := context.TODO()
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer conn.Conn().Close(ctx)

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	// TODO IN THIS COMMIT Modify query to ORDER BY to support descending bool in funciton args
	prefixQuery := `WITH prefix_match(prefix) AS (
		VALUES ('your_prefix_here')
	  )
	  SELECT *
	  FROM your_table
	  WHERE key LIKE (SELECT prefix || '%' FROM prefix_match);
	`
	rows, err := tx.Query(ctx, prefixQuery, prefixKey)
	if err != nil {
		return nil, nil, err
	}

	for rows.Next() {
		var k, v []byte
		if err := rows.Scan(&k, &v); err != nil {
			return nil, nil, err
		}
		keys = append(keys, k)
		values = append(values, v)
		fmt.Printf("key: %+v - value: %+v", k, v)
	}

	return keys, values, nil
}

// Exists ...
func (p *PostgresKV) Exists(key []byte) (bool, error) {
	panic("not implemented") // TODO: Implement
}

// ClearAll ...
func (p *PostgresKV) ClearAll() error {
	panic("not implemented") // TODO: Implement
}
