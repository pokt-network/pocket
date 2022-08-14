package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

const (
	user             = "postgres"
	password         = "secret"
	db               = "postgres"
	sqlSchema        = "test_schema"
	dialect          = "postgres"
	connStringFormat = "postgres://%s:%s@%s/%s?sslmode=disable"
)

func init() {
	PersistenceModule = modules.PersistenceModule(nil) // TODO (team) make thread safe
	PostgresDB = new(persistence.PostgresDB)
}

// TODO (team) cleanup and simplify

var PostgresDB *persistence.PostgresDB
var PersistenceModule modules.PersistenceModule

func SetupPostgresDockerPersistenceMod() (*dockertest.Pool, *dockertest.Resource, modules.PersistenceModule) {
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "12.3",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + db,
		},
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatal(err)
	}
	resource, err := pool.RunWithOptions(&opts, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("***Make sure your docker daemon is running!!*** Could not start resource: %s\n", err.Error())
	}
	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf(connStringFormat, user, password, hostAndPort, db)

	log.Println("Connecting to database on url: ", databaseUrl)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			log.Printf("exit signal %d received\n", sig)
			if err := pool.Purge(resource); err != nil {
				log.Fatalf("could not purge resource: %s", err)
			}
		}
	}()

	resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	cfg := &config.Config{
		GenesisSource: &genesis.GenesisSource{
			Source: &genesis.GenesisSource_Config{
				Config: genesisConfig(),
			},
		},
		Persistence: &config.PersistenceConfig{
			PostgresUrl:    databaseUrl,
			NodeSchema:     sqlSchema,
			BlockStorePath: "",
		},
	}
	err = cfg.HydrateGenesisState()
	if err != nil {
		log.Fatalf("could not hydrate genesis state during postgres setup: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	poolRetryChan := make(chan struct{}, 1)
	var persistenceMod modules.PersistenceModule
	if err = pool.Retry(func() error {
		persistenceMod, err = persistence.Create(cfg)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		persistenceMod.Start()
		PersistenceModule = persistenceMod

		ctx, err := persistenceMod.NewRWContext(0)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		PostgresDB.Tx = ctx.(persistence.PostgresContext).DB.Tx

		poolRetryChan <- struct{}{}

		return nil
	}); err != nil {
		log.Fatalf("could not connect to docker: %s", err.Error())
	}

	// Wait for a successful DB connection
	<-poolRetryChan

	return pool, resource, persistenceMod
}

// TODO: Currently exposed only for testing purposes.
func CleanupPostgresDocker(_ *testing.M, pool *dockertest.Pool, resource *dockertest.Resource) {
	// You can't defer this because `os.Exit`` doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}
	os.Exit(0)
}

// TODO(pocket/issues/149): Golang specific solution for teardown
// TODO: Currently exposed only for testing purposes.
func CleanupTest() {
	PostgresDB.Tx.Rollback(context.TODO())
	PersistenceModule.Stop()
}

func genesisConfig() *genesis.GenesisConfig {
	config := &genesis.GenesisConfig{
		NumValidators:   5,
		NumApplications: 1,
		NumFisherman:    1,
		NumServicers:    1,
	}
	return config
}
