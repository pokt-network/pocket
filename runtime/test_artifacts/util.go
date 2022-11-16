package test_artifacts

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/pokt-network/pocket/utility"
)

const (
	user             = "postgres"
	password         = "secret"
	db               = "postgres"
	sqlSchema        = "test_schema"
	dialect          = "postgres"
	connStringFormat = "postgres://%s:%s@%s/%s?sslmode=disable"
)

// DISCUSS(team) both the persistence module and the utility module share this code which is less than ideal
//
//	(see call to action in generator.go to try to remove the cross module testing code)
func SetupPostgresDocker() (*dockertest.Pool, *dockertest.Resource, string) {
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

	poolRetryChan := make(chan struct{}, 1)
	retryConnectFn := func() error {
		_, err := pgx.Connect(context.Background(), databaseUrl)
		if err != nil {
			return fmt.Errorf("unable to connect to database: %v", err)
		}
		poolRetryChan <- struct{}{}
		return nil
	}
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err = pool.Retry(retryConnectFn); err != nil {
		log.Fatalf("could not connect to docker: %s", err.Error())
	}

	// Wait for a successful DB connection
	<-poolRetryChan

	return pool, resource, databaseUrl
}

// TODO(drewsky): Currently exposed only for testing purposes.
func CleanupPostgresDocker(_ *testing.M, pool *dockertest.Pool, resource *dockertest.Resource) {
	// You can't defer this because `os.Exit`` doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}
}

// CLEANUP: Remove this since it's no longer used or necessary.
func CleanupTest(u utility.UtilityContext) {}
