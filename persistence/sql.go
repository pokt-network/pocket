package persistence

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	"log"
	"math/rand"
	"time"
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

func connectAndInitializeDatabase(postgresUrl string, schema string) error {
	ctx := context.TODO()
	// Connect to the DB
	db, err := pgx.Connect(context.Background(), postgresUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	// Create and set schema (see https://github.com/go-pg/pg/issues/351)
	if _, err = db.Exec(ctx, fmt.Sprintf("%s %s", CreateSchemaIfNotExists, schema)); err != nil {
		return err
	}
	if _, err = db.Exec(ctx, fmt.Sprintf("%s %s", SetSearchPathTo, schema)); err != nil {
		return err
	}
	// pgx.MigrateUp(options, "persistence/schema/migrations")
	if _, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, TableName, TableSchema)); err != nil {
		log.Fatalf("Unable to create %s table: %v\n", TableName, err)
	}
	if err := InitializeTables(ctx, db); err != nil {
		log.Fatal(err.Error())
	}
	return nil
}

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
	return nil
}

func InitializeAccountTables(ctx context.Context, db *pgx.Conn) error {
	_, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.AccountTableName, schema.AccountTableSchema))
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.AccountMetaTableName, schema.AccountMetaTableSchema))
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.PoolTableName, schema.PoolTableSchema))
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.PoolMetaTableName, schema.PoolMetaTableSchema))
	if err != nil {
		return err
	}
	return nil
}

func InitializeValidatorTables(ctx context.Context, db *pgx.Conn) error {
	_, err := db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.ValTableName, schema.ValTableSchema))
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.ValMetaTableName, schema.ValMetaTableSchema))
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
	_, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.AppMetaTableName, schema.AppMetaTableSchema))
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
	_, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.FishMetaTableName, schema.FishMetaTableSchema))
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
	_, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.ServiceNodeMetaTableName, schema.ServiceNodeMetaTableSchema))
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, schema.ServiceNodeChainsTableName, schema.ServiceNodeChainsTableSchema))
	if err != nil {
		return err
	}
	return nil
}
