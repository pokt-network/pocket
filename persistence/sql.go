package persistence

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/modules"
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

var _ modules.PersistenceContext = PostgresContext{}

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
		return nil, fmt.Errorf("unable to create %s table: %v", TableName, err)
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
	if err := InitializeAppTables(ctx, db); err != nil {
		return err
	}
	if err := InitializeServiceTables(ctx, db); err != nil {
		return err
	}
	if err := InitializeFishTables(ctx, db); err != nil {
		return err
	}
	if err := InitializeGovTables(ctx, db); err != nil {
		return err
	}
	return nil
}

func InitializeAccountTables(ctx context.Context, db *pgx.Conn) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.AccountTableName, schema.AccountTableSchema)); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, schema.AccountUniqueCreateIndex); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, schema.AccountUniqueDeleteIndex); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.PoolTableName, schema.PoolTableSchema)); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, schema.PoolUniqueCreateIndex); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, schema.PoolUniqueDeleteIndex); err != nil {
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

func InitializeAppTables(ctx context.Context, db *pgx.Conn) error {
	_, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.AppTableName, schema.AppTableSchema))
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.AppChainsTableName, schema.AppChainsTableSchema))
	if err != nil {
		return err
	}
	return nil
}

func InitializeFishTables(ctx context.Context, db *pgx.Conn) error {
	_, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.FishTableName, schema.FishTableSchema))
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.FishChainsTableName, schema.FishChainsTableSchema))
	if err != nil {
		return err
	}
	return nil
}

func InitializeServiceTables(ctx context.Context, db *pgx.Conn) error {
	_, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.ServiceNodeTableName, schema.ServiceNodeTableSchema))
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.ServiceNodeChainsTableName, schema.ServiceNodeChainsTableSchema))
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

func (p PostgresContext) ClearAllDebug() error {
	ctx, conn, err := p.DB.GetCtxAndConnection()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.ClearAllValidatorsQuery()); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.ClearAllFishermanQuery()); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.ClearAllFishermanChainsQuery()); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.ClearAllServiceNodesChainsQuery()); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.ClearAllServiceNodesQuery()); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.ClearAllAppQuery()); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.ClearAllAppChainsQuery()); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.ClearAllGovQuery()); err != nil {
		return err
	}
	return nil
}
