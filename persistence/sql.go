package persistence

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	PostgresSchemaEnvVar = "POSTGRES_SCHEMA"

	CreateSchemaIfNotExists = "CREATE SCHEMA IF NOT EXISTS"
	SetSearchPathTo         = "SET search_path TO"
	CreateTableIfNotExists  = "CREATE TABLE IF NOT EXISTS"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func connectAndInitializeDatabase(postgresUrl string) error {
	ctx := context.TODO()
	conn, err := pgx.Connect(context.Background(), postgresUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	schema := os.Getenv(PostgresSchemaEnvVar)
	if _, err = conn.Exec(ctx, fmt.Sprintf("%s %s", CreateSchemaIfNotExists, schema)); err != nil {
		return err
	}
	if _, err = conn.Exec(ctx, fmt.Sprintf("%s %s", SetSearchPathTo, schema)); err != nil {
		return err
	}
	return nil
}
