package persistence

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v4"
)

const (
	PostgresSchemaEnvVar = "POSTGRES_SCHEMA"

	CreateSchemaIfNotExists = "CREATE SCHEMA IF NOT EXISTS"
	SetSearchPathTo         = "SET search_path TO"
	CreateTableIfNotExists  = "CREATE TABLE IF NOT EXISTS"
	TableName               = "users"
	TableSchema             = "(id int)"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// TODO(drewsky): Rethink how `connectAndInitializeDatabase` should be implemented in small
// subcomponents, but this curent implementation is more than sufficient for the time being.
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
	if _, err = db.Exec(ctx, fmt.Sprintf(`%s %s %s`, CreateTableIfNotExists, TableName, TableSchema)); err != nil {
		log.Fatalf("Unable to create %s table: %v\n", TableName, err)
	}

	// TODO(olshansky;github.com/pokt-network/pocket/issues/77): Enable proper up and down migrations
	// pgx.MigrateUp(options, "persistence/schema/migrations")

	return nil
}
