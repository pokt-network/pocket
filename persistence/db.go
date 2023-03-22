package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/runtime/configs"
)

const (
	CreateSchema    = "CREATE SCHEMA"
	SetSearchPathTo = "SET search_path TO"
	CreateTable     = "CREATE TABLE"

	IfNotExists = "IF NOT EXISTS"

	CreateEnumType = "CREATE TYPE %s AS ENUM"

	// DUPLICATE OBJECT error. For reference: https://www.postgresql.org/docs/8.4/errcodes-appendix.html
	DuplicateObjectErrorCode = "42710"

	// TODO: Make this a node configuration
	connTimeout = 5 * time.Second
)

// TODO: Move schema related functionality into its own package
var protocolActorSchemas = []types.ProtocolActorSchema{
	types.ApplicationActor,
	types.FishermanActor,
	types.ServicerActor,
	types.ValidatorActor,
}

func (pg *PostgresContext) getCtxAndTx() (context.Context, pgx.Tx) {
	return context.TODO(), pg.tx
}

func initializePool(cfg *configs.PersistenceConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.GetPostgresUrl())
	if err != nil {
		return nil, fmt.Errorf("unable to create database config: %v", err)
	}

	config.MaxConns = cfg.GetMaxConnsCount()
	config.MinConns = cfg.GetMinConnsCount()

	maxConnLifetime, err := time.ParseDuration(cfg.GetMaxConnLifetime())
	if err != nil {
		return nil, fmt.Errorf("unable to set max connection lifetime: %v", err)
	}
	config.MaxConnLifetime = maxConnLifetime

	maxConnIdleTime, err := time.ParseDuration(cfg.GetMaxConnIdleTime())
	if err != nil {
		return nil, fmt.Errorf("unable to set max connection idle time : %v", err)
	}
	config.MaxConnIdleTime = maxConnIdleTime

	healthCheckPeriod, err := time.ParseDuration(cfg.GetHealthCheckPeriod())
	if err != nil {
		return nil, fmt.Errorf("unable to set healthcheck period: %v", err)
	}
	config.HealthCheckPeriod = healthCheckPeriod

	// Update the base connection configurations
	config.ConnConfig.ConnectTimeout = connTimeout

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return pool, nil
}

func connectToPool(pool *pgxpool.Pool, nodeSchema string) (*pgxpool.Conn, error) {
	ctx := context.TODO()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to acquire connection from pool: %v", err)
	}

	// Creating and setting a new schema so we can running multiple nodes on one postgres instance. See
	// more details at https://github.com/go-pg/pg/issues/351.
	if _, err = conn.Exec(ctx, fmt.Sprintf("%s %s %s", CreateSchema, IfNotExists, nodeSchema)); err != nil {
		return nil, err
	}

	// Creating and setting a new schema so we can run multiple nodes on one postgres instance.
	// See more details at https://github.com/go-pg/pg/issues/351.
	if _, err := conn.Exec(ctx, fmt.Sprintf("%s %s %s", CreateSchema, IfNotExists, nodeSchema)); err != nil {
		return nil, err
	}
	if _, err := conn.Exec(ctx, fmt.Sprintf("%s %s", SetSearchPathTo, nodeSchema)); err != nil {
		return nil, err
	}

	return conn, nil
}

// TODO(#77): Enable proper up and down migrations
func initializeDatabase(conn *pgxpool.Conn) error {
	// Initialize the tables if they don't already exist
	if err := initializeAllTables(context.TODO(), conn); err != nil {
		return fmt.Errorf("unable to initialize tables: %v", err)
	}
	return nil
}

// TODO(#77): Delete all the `initializeAllTables` calls once proper migrations are implemented.
func initializeAllTables(ctx context.Context, db *pgxpool.Conn) error {
	if err := initializeAccountTables(ctx, db); err != nil {
		return err
	}

	if err := initializeGovTables(ctx, db); err != nil {
		return err
	}

	if err := initializeBlockTables(ctx, db); err != nil {
		return err
	}

	for _, actor := range protocolActorSchemas {
		if err := initializeProtocolActorTables(ctx, db, actor); err != nil {
			return err
		}
	}

	return nil
}

func initializeProtocolActorTables(ctx context.Context, db *pgxpool.Conn, actor types.ProtocolActorSchema) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, actor.GetTableName(), actor.GetTableSchema())); err != nil {
		return err
	}
	if actor.GetChainsTableName() != "" {
		if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, actor.GetChainsTableName(), actor.GetChainsTableSchema())); err != nil {
			return err
		}
	}
	return nil
}

func initializeAccountTables(ctx context.Context, db *pgxpool.Conn) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, types.AccountTableName, types.Account.GetTableSchema())); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, types.PoolTableName, types.Pool.GetTableSchema())); err != nil {
		return err
	}
	return nil
}

func initializeGovTables(ctx context.Context, db *pgxpool.Conn) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s`, fmt.Sprintf(CreateEnumType, types.ValTypeName), types.ValTypeEnumTypes)); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code != DuplicateObjectErrorCode {
			return err
		}
	}

	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, types.ParamsTableName, types.ParamsTableSchema)); err != nil {
		return err
	}

	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, types.FlagsTableName, types.FlagsTableSchema)); err != nil {
		return err
	}

	return nil
}

func initializeBlockTables(ctx context.Context, db *pgxpool.Conn) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, types.BlockTableName, types.BlockTableSchema)); err != nil {
		return err
	}
	return nil
}
