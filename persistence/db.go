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
)

// TODO: Move schema related functionality into its own package
var protocolActorSchemas = []types.ProtocolActorSchema{
	types.ApplicationActor,
	types.FishermanActor,
	types.ServicerActor,
	types.ValidatorActor,
}

func (pg *PostgresContext) getCtxAndTx() (context.Context, pgx.Tx) {
	return context.TODO(), pg.getTx()
}

func (pg *PostgresContext) getTx() pgx.Tx {
	return pg.tx
}

func (pg *PostgresContext) ResetContext() error {
	if pg == nil {
		return nil
	}
	tx := pg.getTx()
	if tx == nil {
		return nil
	}
	conn := tx.Conn()
	if conn == nil {
		return nil
	}
	if !conn.IsClosed() {
		if err := pg.Release(); err != nil {
			pg.logger.Error().Err(err).Bool("TODO", true).Msg("error releasing write context")
		}
	}
	pg.tx = nil
	return nil
}

func connectToDatabase(cfg *configs.PersistenceConfig) (*pgx.Conn, error) {
	ctx := context.TODO()

	config, err := pgxpool.ParseConfig(cfg.GetPostgresUrl())
	if err != nil {
		return nil, fmt.Errorf("unable to create database config: %v", err)
	}
	maxConnLifetime, err := time.ParseDuration(cfg.GetMaxConnLifetime())
	if err == nil {
		config.MaxConnLifetime = maxConnLifetime
	} else {
		return nil, fmt.Errorf("unable to set max connection lifetime: %v", err)
	}
	maxConnIdleTime, err := time.ParseDuration(cfg.GetMaxConnIdleTime())
	if err == nil {
		config.MaxConnIdleTime = maxConnIdleTime
	} else {
		return nil, fmt.Errorf("unable to set max connection idle time : %v", err)
	}
	config.MaxConns = cfg.GetMaxConnsCount()
	config.MinConns = cfg.GetMinConnsCount()
	healthCheckPeriod, err := time.ParseDuration(cfg.GetHealthCheckPeriod())
	if err == nil {
		config.HealthCheckPeriod = healthCheckPeriod
	} else {
		return nil, fmt.Errorf("unable to set healthcheck period: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	conn, _ := pool.Acquire(ctx)

	nodeSchema := cfg.GetNodeSchema()
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

	return conn.Conn(), nil
}

// TODO(pokt-network/pocket/issues/77): Enable proper up and down migrations
func initializeDatabase(conn *pgx.Conn) error {
	// Initialize the tables if they don't already exist
	if err := initializeAllTables(context.TODO(), conn); err != nil {
		return fmt.Errorf("unable to initialize tables: %v", err)
	}
	return nil
}

// TODO(pokt-network/pocket/issues/77): Delete all the `initializeAllTables` calls once proper migrations are implemented.
func initializeAllTables(ctx context.Context, db *pgx.Conn) error {
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

func initializeProtocolActorTables(ctx context.Context, db *pgx.Conn, actor types.ProtocolActorSchema) error {
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

func initializeAccountTables(ctx context.Context, db *pgx.Conn) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, types.AccountTableName, types.Account.GetTableSchema())); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, types.PoolTableName, types.Pool.GetTableSchema())); err != nil {
		return err
	}
	return nil
}

func initializeGovTables(ctx context.Context, db *pgx.Conn) error {
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

func initializeBlockTables(ctx context.Context, db *pgx.Conn) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, types.BlockTableName, types.BlockTableSchema)); err != nil {
		return err
	}
	return nil
}
