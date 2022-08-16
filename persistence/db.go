package persistence

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/modules"
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

var _ modules.PersistenceRWContext = &PostgresContext{}

// TODO: These are only externalized for testing purposes, so they should be made private and
//
//	it is trivial to create a helper to initial a context with some values.
type PostgresContext struct {
	Height int64
	DB     PostgresDB
}

type PostgresDB struct {
	Tx         pgx.Tx
	Blockstore kvstore.KVStore
}

func (pg *PostgresDB) GetCtxAndTxn() (context.Context, pgx.Tx, error) {
	tx, err := pg.GetTxn()
	// IMPROVE: Depending on how the use of `PostgresContext` evolves, we may be able to get
	// access to these directly via the postgres module.
	//PostgresDB *pgx.Conn
	//BlockStore kvstore.KVStore
	return context.TODO(), tx, err
}

func (pg *PostgresDB) GetTxn() (pgx.Tx, error) {
	return pg.Tx, nil
}

func (pg *PostgresContext) GetContext() (context.Context, error) {
	return context.TODO(), nil
}

var protocolActorSchemas = []schema.ProtocolActorSchema{
	schema.ApplicationActor,
	schema.FishermanActor,
	schema.ServiceNodeActor,
	schema.ValidatorActor,
}

// TODO(pokt-network/pocket/issues/77): Enable proper up and down migrations
// TODO: Split `connect` and `initialize` into two separate compnents
func ConnectAndInitializeDatabase(postgresUrl string, schema string) (*pgx.Conn, error) {
	ctx := context.TODO()

	// Connect to the DB
	db, err := pgx.Connect(context.Background(), postgresUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	// Creating and setting a new schema so we can running multiple nodes on one postgres instance. See
	// more details at https://github.com/go-pg/pg/issues/351.
	if _, err = db.Exec(ctx, fmt.Sprintf("%s %s %s", CreateSchema, IfNotExists, schema)); err != nil {
		return nil, err
	}

	if _, err = db.Exec(ctx, fmt.Sprintf("%s %s", SetSearchPathTo, schema)); err != nil {
		return nil, err
	}

	if err = initializeAllTables(ctx, db); err != nil {
		return nil, fmt.Errorf("unable to initialize tables: %v", err)
	}

	return db, nil

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

func initializeProtocolActorTables(ctx context.Context, db *pgx.Conn, actor schema.ProtocolActorSchema) error {
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
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, schema.AccountTableName, schema.AccountTableSchema)); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, schema.PoolTableName, schema.PoolTableSchema)); err != nil {
		return err
	}
	return nil
}

func initializeGovTables(ctx context.Context, db *pgx.Conn) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s`, fmt.Sprintf(CreateEnumType, schema.ValTypeName), schema.ValTypeEnumTypes)); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code != DuplicateObjectErrorCode {
			return err
		}
	}

	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, schema.ParamsTableName, schema.ParamsTableSchema)); err != nil {
		return err
	}

	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, schema.FlagsTableName, schema.FlagsTableSchema)); err != nil {
		return err
	}

	return nil
}

func initializeBlockTables(ctx context.Context, db *pgx.Conn) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, schema.BlockTableName, schema.BlockTableSchema)); err != nil {
		return err
	}
	return nil
}

// Exposed for debugging purposes only
func (p PostgresContext) ClearAllDebug() error {
	ctx, conn, err := p.DB.GetCtxAndTxn()
	if err != nil {
		return err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	for _, actor := range protocolActorSchemas {
		if _, err = tx.Exec(ctx, actor.ClearAllQuery()); err != nil {
			return err
		}
		if actor.GetChainsTableName() != "" {
			if _, err = tx.Exec(ctx, actor.ClearAllChainsQuery()); err != nil {
				return err
			}
		}
	}

	if _, err = tx.Exec(ctx, schema.ClearAllGovParamsQuery()); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, schema.ClearAllGovFlagsQuery()); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, schema.ClearAllBlocksQuery()); err != nil {
		return err
	}

	return nil
}
