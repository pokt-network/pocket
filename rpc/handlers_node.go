package rpc

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/runtime/configs"
)

// PostV1NodeBackup triggers a backup of the TreeStore, the BlockStore, the PostgreSQL database.
// TECHDEBT: Run each backup process in a goroutine to as elapsed time will become significant
// with the current waterfall approach when even a moderate amount of data resides in each store.
func (s *rpcServer) PostV1NodeBackup(ctx echo.Context) error {
	dir := os.TempDir() // TODO_IN_THIS_COMMIT give this a sane default and make it configurable
	s.logger.Info().Msgf("creating backup in %s", dir)

	// backup the TreeStore
	trees := s.GetBus().GetTreeStore()
	if err := trees.Backup(dir); err != nil {
		return err
	}

	// backup the BlockStore
	path := fmt.Sprintf("%s-blockstore-backup.sql", time.Now().String())
	if err := s.GetBus().GetPersistenceModule().GetBlockStore().Backup(path); err != nil {
		return err
	}

	// backup the Postgres database
	cfg := s.GetBus().GetRuntimeMgr().GetConfig()
	err := postgresBackup(cfg, dir) // TODO_IN_THIS_COMMIT make this point at the right directory per the tests
	if err != nil {
		return err
	}

	s.logger.Info().Msgf("backup created in %s", dir)
	return nil
}

func postgresBackup(cfg *configs.Config, dir string) error {
	filename := fmt.Sprintf("%s-postgres-backup.sql", time.Now().String())
	file, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	// pgurl := cfg.Persistence.PostgresUrl
	// credentials, err := parsePostgreSQLConnectionURL(pgurl)
	if err != nil {
		return err
	}

	cmd := exec.Command("which pg_dump")
	fmt.Printf("cmd.Stdout: %v\n", cmd.Stdout)
	fmt.Printf("cmd.Stderr: %v\n", cmd.Stderr)

	// cmd := exec.Command(fmt.Sprintf("PGPASSWORD=%s", credentials.password), "pg_dump", "-h", credentials.host, "-U", credentials.username, credentials.dbName)
	// cmd.Stdout = file
	// err = cmd.Run()
	// if err != nil {
	// 	return err
	// }

	return nil
}

type credentials struct {
	username string
	password string
	host     string
	dbName   string
	sslMode  string
}

// validate a credentials object for connecting to postgres to create a backup
func parsePostgreSQLConnectionURL(connectionURL string) (*credentials, error) {
	parsedURL, err := url.Parse(connectionURL)
	if err != nil {
		return nil, err
	}

	if parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql" {
		return nil, fmt.Errorf("failed to parse postgres URL")
	}

	username := parsedURL.User.Username()
	password, _ := parsedURL.User.Password()
	host := parsedURL.Host
	dbName := parsedURL.Path[1:] // Remove the leading slash
	query := parsedURL.Query()
	sslMode := query.Get("sslmode")

	return &credentials{
		username: username,
		password: password,
		host:     host,
		dbName:   dbName,
		sslMode:  sslMode,
	}, nil
}
