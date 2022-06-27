package persistence

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
)

const (
	CreateSchemaIfNotExists = "CREATE SCHEMA IF NOT EXISTS"
	SetSearchPathTo         = "SET search_path TO"
	CreateTableIfNotExists  = "CREATE TABLE IF NOT EXISTS"
	TableName               = "users"
	TableSchema             = "(id int)"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//var _ modules.PersistenceContext = PostgresContext{}

type PostgresContext struct {
	Height int64
	DB     PostgresDB
}

type PostgresDB struct {
	Conn *pgx.Conn // TODO (TEAM) use pool of connections
}

func (pg *PostgresDB) GetCtxAndConnection() (context.Context, *pgx.Conn, error) {
	conn, err := pg.GetConnection()
	if err != nil {
		return nil, nil, err
	}
	ctx, err := pg.GetContext()
	if err != nil {
		return nil, nil, err
	}
	return ctx, conn, nil
}

func (pg *PostgresDB) GetConnection() (*pgx.Conn, error) {
	return pg.Conn, nil
}

func (pg *PostgresDB) GetContext() (context.Context, error) {
	return context.TODO(), nil
}

var protocolActorSchemas = []schema.ProtocolActor{
	schema.ApplicationActor,
	schema.FishermanActor,
	schema.ServiceNodeActor,
}

func ConnectAndInitializeDatabase(postgresUrl string, schema string) (*pgx.Conn, error) {
	// TODO(drewsky): Rethink how `connectAndInitializeDatabase` should be implemented in small
	// subcomponents, but this curent implementation is more than sufficient for the time being.
	ctx := context.TODO()
	// Connect to the DB
	db, err := pgx.Connect(context.Background(), postgresUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	// Create and set schema (see https://github.com/go-pg/pg/issues/351)
	if _, err = db.Exec(ctx, fmt.Sprintf("%s %s", CreateSchemaIfNotExists, schema)); err != nil {
		return nil, err
	}
	if _, err = db.Exec(ctx, fmt.Sprintf("%s %s", SetSearchPathTo, schema)); err != nil {
		return nil, err
	}
	if _, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, TableName, TableSchema)); err != nil {
		return nil, fmt.Errorf("unable to create %s table: %v", TableName, err)
	}
	if err := InitializeTables(ctx, db); err != nil {
		return nil, fmt.Errorf("unable to initialize tables: %v", err)
	}
	return db, nil
	// TODO(olshansky;github.com/pokt-network/pocket/issues/77): Enable proper up and down migrations
	// pgx.MigrateUp(options, "persistence/schema/migrations")
}

// TODO(olshansky;github.com/pokt-network/pocket/issues/77): Delete all the `InitializeTables` calls
// once proper migrations are implemented.
func InitializeTables(ctx context.Context, db *pgx.Conn) error {
	if err := InitializeAccountTables(ctx, db); err != nil {
		return err
	}
	if err := InitializeValidatorTables(ctx, db); err != nil {
		return err
	}
	if err := InitializeGovTables(ctx, db); err != nil {
		return err
	}
	for _, actor := range protocolActorSchemas {
		InitializeProtocolActorTables(ctx, db, actor)

	}
	return nil
}

func InitializeProtocolActorTables(ctx context.Context, db *pgx.Conn, actor schema.ProtocolActor) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, actor.GetTableName(), actor.GetTableSchema())); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, actor.GetChainsTableName(), actor.GetChainsTableSchema())); err != nil {
		return err
	}
	return nil
}

func InitializeAccountTables(ctx context.Context, db *pgx.Conn) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.AccountTableName, schema.AccountTableSchema)); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.PoolTableName, schema.PoolTableSchema)); err != nil {
		return err
	}
	return nil
}

func InitializeValidatorTables(ctx context.Context, db *pgx.Conn) error {
	_, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.ValTableName, schema.ValTableSchema))
	if err != nil {
		return err
	}
	return nil
}

func InitializeGovTables(ctx context.Context, db *pgx.Conn) error {
	_, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.ParamsTableName, schema.ParamsTableSchema))
	if err != nil {
		return err
	}
	return nil
}

// Only exposed for testing purposes.

var clearFunctions = []func() string{
	schema.ClearAllValidatorsQuery,
	schema.ClearAllGovQuery,
}

func (p PostgresContext) ClearAllDebug() error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	for _, clearFunc := range clearFunctions {
		if _, err = tx.Exec(ctx, clearFunc()); err != nil {
			return err
		}
	}

	for _, actor := range protocolActorSchemas {
		if _, err = tx.Exec(ctx, actor.ClearAllQuery()); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, actor.ClearAllChainsQuery()); err != nil {
			return err
		}
	}
	return nil
}
