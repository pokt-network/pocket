package test

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"testing"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/types"
)

const (
	user             = "postgres"
	password         = "secret"
	db               = "postgres"
	schema           = "test_schema"
	localhost        = "0.0.0.0"
	port             = "5432"
	dialect          = "postgres"
	connStringFormat = "postgres://%s:%s@localhost:%s/%s?sslmode=disable"
)

var (
	DefaultChains         = []string{"0001"}
	ChainsToUpdate        = []string{"0002"}
	DefaultServiceUrl     = "https://foo.bar"
	DefaultPoolName       = "TESTING_POOL"
	DefaultDeltaBig       = big.NewInt(100)
	DefaultAccountBig     = big.NewInt(1000000)
	DefaultStakeBig       = big.NewInt(1000000000000000)
	DefaultMaxRelaysBig   = big.NewInt(1000000)
	DefaultDeltaAmount    = types.BigIntToString(DefaultDeltaBig)
	DefaultAccountAmount  = types.BigIntToString(DefaultAccountBig)
	DefaultStake          = types.BigIntToString(DefaultStakeBig)
	DefaultMaxRelays      = types.BigIntToString(DefaultMaxRelaysBig)
	StakeToUpdate         = types.BigIntToString((&big.Int{}).Add(DefaultStakeBig, DefaultDeltaBig))
	ParamToUpdate         = 2
	DefaultAccountBalance = DefaultStake
	DefaultStakeStatus    = persistence.StakedStatus
	// DISCUSS(drewsky): Not a fan of using `Default` as something that has semantic meaning (i.e. currently active). Pick a better name together.
	DefaultPauseHeight     = int64(-1)
	DefaultUnstakingHeight = int64(-1)
	PauseHeightToSet       = 1
)

// TODO:(team) make these tests thread safe
var PostgresDB *persistence.PostgresDB

func init() {
	PostgresDB = new(persistence.PostgresDB)
}

func TestMain(m *testing.M) {
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "12.3",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + db,
		},
		ExposedPorts: []string{port},
		PortBindings: map[docker.Port][]docker.PortBinding{
			port: {
				{HostIP: localhost, HostPort: port},
			},
		},
	}
	connString := fmt.Sprintf(connStringFormat, user, password, port, db)

	defer func() {
		ctx, _ := PostgresDB.GetContext()
		PostgresDB.Conn.Close(ctx)
		ctx.Done()
	}()

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&opts)
	if err != nil {
		log.Fatalf("***Make sure your docker daemon is running!!*** Could not start resource: %s\n", err.Error())
	}

	// DISCUSS(drewsky): Is this some sort of cleanup?
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill) // Why not syscall.SIGTERM?
	go func() {
		for sig := range c {
			log.Printf("exit signal %d received\n", sig)
			if err := pool.Purge(resource); err != nil {
				log.Fatalf("could not purge resource: %s", err)
			}
		}
	}()

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err = pool.Retry(func() error {
		conn, err := persistence.ConnectAndInitializeDatabase(connString, schema)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		PostgresDB.Conn = conn
		return nil
	}); err != nil {
		log.Fatalf("could not connect to docker: %s", err.Error())
	}
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}
	os.Exit(code)
}
