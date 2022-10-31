package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/types"
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

var protocolActorSchemas = []types.ProtocolActorSchema{
	types.ApplicationActor,
	types.FishermanActor,
	types.ServiceNodeActor,
	types.ValidatorActor,
}

// A list of functions to clear data from the DB not associated with protocol actors
var nonActorClearFunctions = []func() string{
	types.ClearAllAccounts,
	types.ClearAllPools,
	types.ClearAllGovParamsQuery,
	types.ClearAllGovFlagsQuery,
	types.ClearAllBlocksQuery,
}

var _ modules.PersistenceRWContext = &PostgresContext{}

type PostgresContext struct {
	Height int64 // TODO(olshansky): `Height` is only externalized for testing purposes. Replace with helpers...
	conn   *pgx.Conn
	tx     pgx.Tx

	// TODO(#315): Should be
	currentStateHash []byte

	// REFACTOR_IN_THIS_COMMIT: Access `blockStore` and `merkleTree` from the persistence module via bus.
	blockStore kvstore.KVStore
	txIndexer  indexer.TxIndexer
	stateTrees *stateTrees
}

func (pg *PostgresContext) GetCtxAndTx() (context.Context, pgx.Tx, error) {
	return context.TODO(), pg.GetTx(), nil
}

func (pg *PostgresContext) GetTx() pgx.Tx {
	return pg.tx
}

func (pg *PostgresContext) GetCtx() (context.Context, error) {
	return context.TODO(), nil
}

// TECHDEBT: Implement proper connection pooling
func connectToDatabase(postgresUrl string, schema string) (*pgx.Conn, error) {
	ctx := context.TODO()

	conn, err := pgx.Connect(context.Background(), postgresUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	// Creating and setting a new schema so we can running multiple nodes on one postgres instance. See
	// more details at https://github.com/go-pg/pg/issues/351.
	if _, err = conn.Exec(ctx, fmt.Sprintf("%s %s %s", CreateSchema, IfNotExists, schema)); err != nil {
		return nil, err
	}

	// Creating and setting a new schema so we can run multiple nodes on one postgres instance.
	// See more details at https://github.com/go-pg/pg/issues/351.
	if _, err := conn.Exec(ctx, fmt.Sprintf("%s %s %s", CreateSchema, IfNotExists, schema)); err != nil {
		return nil, err
	}
	if _, err := conn.Exec(ctx, fmt.Sprintf("%s %s", SetSearchPathTo, schema)); err != nil {
		return nil, err
	}

	return conn, nil
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
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, types.AccountTableName, types.AccountTableSchema)); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s %s`, CreateTable, IfNotExists, types.PoolTableName, types.PoolTableSchema)); err != nil {
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

// Exposed for testing purposes only
func (p *PostgresContext) DebugClearAll() error {
	ctx, clearTx, err := p.GetCtxAndTx()
	if err != nil {
		return err
	}

	for _, actor := range protocolActorSchemas {
		if _, err = clearTx.Exec(ctx, actor.ClearAllQuery()); err != nil {
			return err
		}
		if actor.GetChainsTableName() != "" {
			if _, err = clearTx.Exec(ctx, actor.ClearAllChainsQuery()); err != nil {
				return err
			}
		}
	}

	for _, clearFn := range nonActorClearFunctions {
		if _, err := clearTx.Exec(ctx, clearFn()); err != nil {
			return err
		}
	}

	if err = clearTx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
