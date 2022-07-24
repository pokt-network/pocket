package test

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"golang.org/x/exp/slices"

	"github.com/stretchr/testify/require"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
)

const (
	user             = "postgres"
	password         = "secret"
	db               = "postgres"
	sql_schema       = "test_schema"
	port             = "5431" // Intentionally not `5432` so as not to interfere with local settings
	dialect          = "postgres"
	connStringFormat = "postgres://%s:%s@%s/%s?sslmode=disable"
)

var (
	DefaultChains     = []string{"0001"}
	ChainsToUpdate    = []string{"0002"}
	DefaultServiceUrl = "https://foo.bar"
	DefaultPoolName   = "TESTING_POOL"

	DefaultDeltaBig     = big.NewInt(100)
	DefaultAccountBig   = big.NewInt(1000000)
	DefaultStakeBig     = big.NewInt(1000000000000000)
	DefaultMaxRelaysBig = big.NewInt(1000000)

	DefaultAccountAmount = types.BigIntToString(DefaultAccountBig)
	DefaultStake         = types.BigIntToString(DefaultStakeBig)
	DefaultMaxRelays     = types.BigIntToString(DefaultMaxRelaysBig)
	StakeToUpdate        = types.BigIntToString((&big.Int{}).Add(DefaultStakeBig, DefaultDeltaBig))

	DefaultStakeStatus     = persistence.StakedStatus
	DefaultPauseHeight     = int64(-1)
	DefaultUnstakingHeight = int64(-1)
)

var PostgresDB *persistence.PostgresDB

// TODO(team): make these tests thread safe
func init() {
	PostgresDB = new(persistence.PostgresDB)
}

// See https://github.com/ory/dockertest as reference for the template of this code
// Postgres example can be found here: https://github.com/ory/dockertest/blob/v3/examples/PostgreSQL.md
func TestMain(m *testing.M) {
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "12.3",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + db,
		},
	}

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
	resource, err := pool.RunWithOptions(&opts, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("***Make sure your docker daemon is running!!*** Could not start resource: %s\n", err.Error())
	}

	// Example: https://github.com/ory/dockertest/blob/v3/examples/PostgreSQL.md
	// Reasoning: https://github.com/ory/dockertest/blob/v3/examples/PostgreSQL.md
	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf(connStringFormat, user, password, hostAndPort, db)

	log.Println("Connecting to database on url: ", databaseUrl)

	// DOCUMENT: Why do we not call `syscall.SIGTERM` here
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

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err = pool.Retry(func() error {
		conn, err := persistence.ConnectAndInitializeDatabase(databaseUrl, sql_schema)
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

	// You can't defer this because `os.Exit`` doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}
	os.Exit(code)
}

// IMPROVE(team): Extend this to more complex and variable test cases challenging & randomizing the state of persistence.
func fuzzSingleProtocolActor(
	f *testing.F,
	newTestActor func() (schema.BaseActor, error),
	getTestActor func(db persistence.PostgresContext, address string) (*schema.BaseActor, error),
	protocolActorSchema schema.ProtocolActorSchema) {

	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}

	err := db.ClearAllDebug()
	require.NoError(f, err)

	actor, err := newTestActor()
	require.NoError(f, err)

	err = db.InsertActor(protocolActorSchema, actor)
	require.NoError(f, err)

	// IMPROVE(team): Extend this to make sure we have full code coverage of the persistence context operations.
	operations := []string{
		"UpdateActor",

		"GetActorsReadyToUnstake",
		"GetActorStatus",
		"GetActorPauseHeight",
		"GetActorOutputAddr",

		"SetActorUnstakingHeight",
		"SetActorPauseHeight",
		"SetPausedActorToUnstaking",

		"IncrementHeight"}
	numOperationTypes := len(operations)

	numDbOperations := 100
	for i := 0; i < numDbOperations; i++ {
		f.Add(operations[rand.Intn(numOperationTypes)])
	}

	f.Fuzz(func(t *testing.T, op string) {
		originalActor, err := getTestActor(db, actor.Address)
		require.NoError(t, err)

		addr, err := hex.DecodeString(originalActor.Address)
		require.NoError(t, err)

		switch op {
		case "UpdateActor":
			numParamUpdatesSupported := 3
			newStakedTokens := originalActor.StakedTokens
			newChains := originalActor.Chains
			newActorSpecificParam := originalActor.ActorSpecificParam

			iterations := rand.Intn(2)
			for i := 0; i < iterations; i++ {
				switch rand.Intn(numParamUpdatesSupported) {
				case 0:
					newStakedTokens = getRandomBigIntString()
				case 1:
					switch protocolActorSchema.GetActorSpecificColName() {
					case schema.ServiceURLCol:
						newActorSpecificParam = getRandomServiceURL()
					case schema.MaxRelaysCol:
						newActorSpecificParam = getRandomBigIntString()
					default:
						t.Error("Unexpected actor specific column name")
					}
				case 2:
					if protocolActorSchema.GetChainsTableName() != "" {
						newChains = getRandomChains()
					}
				}
			}
			updatedActor := schema.BaseActor{
				Address:            originalActor.Address,
				PublicKey:          originalActor.PublicKey,
				StakedTokens:       newStakedTokens,
				ActorSpecificParam: newActorSpecificParam,
				OutputAddress:      originalActor.OutputAddress,
				PausedHeight:       originalActor.PausedHeight,
				UnstakingHeight:    originalActor.UnstakingHeight,
				Chains:             newChains,
			}
			err = db.UpdateActor(protocolActorSchema, updatedActor)
			require.NoError(t, err)

			newActor, err := getTestActor(db, originalActor.Address)
			require.NoError(t, err)

			require.ElementsMatch(t, newActor.Chains, newChains, "staked chains not updated")
			require.Equal(t, newActor.StakedTokens, newStakedTokens, "staked tokens not updated")
			require.Equal(t, newActor.ActorSpecificParam, newActorSpecificParam, "actor specific param not updated")
		case "GetActorsReadyToUnstake":
			unstakingActors, err := db.GetActorsReadyToUnstake(protocolActorSchema, db.Height)
			require.NoError(t, err)

			if originalActor.UnstakingHeight != db.Height { // Not ready to unstake
				require.Nil(t, unstakingActors)
			} else {
				idx := slices.IndexFunc(unstakingActors, func(a *types.UnstakingActor) bool {
					return originalActor.Address == hex.EncodeToString(a.Address)
				})
				require.NotEqual(t, idx, -1, fmt.Sprintf("actor that is unstaking was not found %+v", originalActor))
			}
		case "GetActorStatus":
			status, err := db.GetActorStatus(protocolActorSchema, addr, db.Height)
			require.NoError(t, err)

			switch {
			case originalActor.UnstakingHeight == DefaultUnstakingHeight:
				require.Equal(t, persistence.StakedStatus, status, "actor status should be staked")
			case originalActor.UnstakingHeight > db.Height:
				require.Equal(t, persistence.UnstakingStatus, status, "actor status should be unstaking")
			default:
				require.Equal(t, persistence.UnstakedStatus, status, "actor status should be unstaked")
			}
		case "GetActorPauseHeight":
			pauseHeight, err := db.GetActorPauseHeightIfExists(protocolActorSchema, addr, db.Height)
			require.NoError(t, err)

			require.Equal(t, originalActor.PausedHeight, pauseHeight, "pause height incorrect")
		case "SetActorUnstakingHeight":
			newUnstakingHeight := rand.Int63()

			err = db.SetActorUnstakingHeightAndStatus(protocolActorSchema, addr, newUnstakingHeight)
			require.NoError(t, err)

			newActor, err := getTestActor(db, originalActor.Address)
			require.NoError(t, err)

			require.Equal(t, newUnstakingHeight, newActor.UnstakingHeight, "setUnstakingHeight")
		case "SetActorPauseHeight":
			newPauseHeight := rand.Int63()

			err = db.SetActorPauseHeight(protocolActorSchema, addr, newPauseHeight)
			require.NoError(t, err)

			newActor, err := getTestActor(db, actor.Address)
			require.NoError(t, err)

			require.Equal(t, newPauseHeight, newActor.PausedHeight, "setPauseHeight")
		case "SetPausedActorToUnstaking":
			newUnstakingHeight := db.Height + int64(rand.Intn(15))
			err = db.SetActorStatusAndUnstakingHeightIfPausedBefore(protocolActorSchema, db.Height, newUnstakingHeight)
			require.NoError(t, err)

			newActor, err := getTestActor(db, originalActor.Address)
			require.NoError(t, err)

			if db.Height > originalActor.PausedHeight { // isPausedAndReadyToUnstake
				require.Equal(t, newActor.UnstakingHeight, newUnstakingHeight, "setPausedToUnstaking")
			}
		case "GetActorOutputAddr":
			outputAddr, err := db.GetActorOutputAddress(protocolActorSchema, addr, db.Height)
			require.NoError(t, err)

			require.Equal(t, originalActor.OutputAddress, hex.EncodeToString(outputAddr), "output address incorrect")
		case "IncrementHeight":
			db.Height++
		default:
			t.Errorf("Unexpected operation fuzzing operation %s", op)
		}
	})
}

func getRandomChains() (chains []string) {
	setRandomSeed()

	charOptions := "ABCDEF0123456789"
	numCharOptions := len(charOptions)

	chainsMap := make(map[string]struct{})
	for i := 0; i < rand.Intn(14)+1; i++ {
		b := make([]byte, 4)
		for i := range b {
			b[i] = charOptions[rand.Intn(numCharOptions)]
		}
		if _, found := chainsMap[string(b)]; found {
			i--
			continue
		}
		chainsMap[string(b)] = struct{}{}
		chains = append(chains, string(b))
	}
	return
}

func getRandomServiceURL() string {
	setRandomSeed()

	charOptions := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numCharOptions := len(charOptions)

	b := make([]byte, rand.Intn(12))
	for i := range b {
		b[i] = charOptions[rand.Intn(numCharOptions)]
	}

	return fmt.Sprintf("https://%s.com", string(b))
}

func getRandomBigIntString() string {
	return types.BigIntToString(big.NewInt(rand.Int63()))
}

func setRandomSeed() {
	rand.Seed(time.Now().UnixNano())
}
