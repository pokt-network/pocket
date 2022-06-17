package test

import (
	"encoding/hex"
	"fmt"
	schema2 "github.com/pokt-network/pocket/persistence/schema"
	"github.com/stretchr/testify/require"
	"log"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"testing"
	"time"

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
	DefaultChains        = []string{"0001"}
	ChainsToUpdate       = []string{"0002"}
	DefaultServiceUrl    = "https://foo.bar"
	DefaultPoolName      = "TESTING_POOL"
	DefaultDeltaBig      = big.NewInt(100)
	DefaultAccountBig    = big.NewInt(1000000)
	DefaultStakeBig      = big.NewInt(1000000000000000)
	DefaultMaxRelaysBig  = big.NewInt(1000000)
	DefaultAccountAmount = types.BigIntToString(DefaultAccountBig)
	DefaultStake         = types.BigIntToString(DefaultStakeBig)
	DefaultMaxRelays     = types.BigIntToString(DefaultMaxRelaysBig)
	StakeToUpdate        = types.BigIntToString((&big.Int{}).Add(DefaultStakeBig, DefaultDeltaBig))
	ParamToUpdate        = 2
	DefaultStakeStatus   = persistence.StakedStatus
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

func GetRandomServiceURL() string {
	rand.Seed(time.Now().UnixNano())
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, rand.Intn(12))
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return "https://" + string(b) + ".com"
}

func fuzzActor(f *testing.F, newActor func() (schema2.GenericActor, error),
	insertActorQuery func(string, string, string, string, string, int64, int64, []string, int64) string,
	getTestActor func(db persistence.PostgresContext, address string) (*schema2.GenericActor, error),
	hasServiceURL bool,
	updateActorQuery func(string, string, string, int64) string,
	updateChainsQuery func(string, []string, int64) string,
	chainsTableName string,
	readyToUnstakeQuery func(int64) string,
	actorStatusQuery func(string, int64) string,
	actorPauseHeightIfExists func(string, int64) string,
	getActorQuery func(string, int64) string,
	getChainsQuery func(string, int64) string,
	setActorUnstakingHeight func(string, int64, int64) string,
	setActorPauseHeightQuery func(string, int64, int64) string,
	setActorUnstakingPausedBefore func(int64, int64, int64) string,
	getActorOutputAddress func(string, int64) string) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	err := db.ClearAllDebug()
	if err != nil {
		panic(err)
	}
	ops := []string{"Update", "GetReadyToUnstake",
		"GetStatus", "GetPauseHeight", "SetUnstakingHeight", "SetPauseHeight",
		"SetPausedToUnstaking", "GetOutput", "NextHeight"}
	actor, err := newActor()
	if err != nil {
		panic(err)
	}
	err = db.InsertActor(actor, insertActorQuery)
	if err != nil {
		panic(err)
	}
	numOptions := len(ops)
	numOperations := 100
	for i := 0; i < numOperations; i++ {
		f.Add(ops[rand.Intn(numOptions)])
	}
	f.Fuzz(func(t *testing.T, op string) {
		switch op {
		case "Update":
			originalActor, err := getTestActor(db, actor.Address)
			require.NoError(t, err)
			newStakedTokens := originalActor.StakedTokens
			newChains := originalActor.Chains
			genericParam := originalActor.GenericParam
			iterations := rand.Intn(2)
			for i := 0; i < iterations; i++ {
				switch rand.Intn(3) {
				case 0:
					newStakedTokens = types.BigIntToString(big.NewInt(rand.Int63()))
				case 1:
					if hasServiceURL {
						genericParam = GetRandomServiceURL()
					} else {
						genericParam = types.BigIntToString(big.NewInt(rand.Int63()))
					}
				case 2:
					if chainsTableName != "" {
						newChains = GetRandomChains()
					}
				default: // do nothing
				}
			}
			updatedActor := schema2.GenericActor{
				Address:         originalActor.Address,
				PublicKey:       originalActor.PublicKey,
				StakedTokens:    newStakedTokens,
				GenericParam:    genericParam,
				OutputAddress:   originalActor.OutputAddress,
				PausedHeight:    originalActor.PausedHeight,
				UnstakingHeight: originalActor.UnstakingHeight,
				Chains:          newChains,
			}
			err = db.UpdateActor(updatedActor, updateActorQuery, updateChainsQuery, chainsTableName)
			require.NoError(t, err)
			nActor, err := getTestActor(db, originalActor.Address)
			require.NoError(t, err)
			require.Equal(t, nActor.GenericParam, genericParam, "update maxRelays")
			require.True(t, isUnorderedEqual(nActor.Chains, newChains), "update newChains: "+fmt.Sprintf("%s, %s", nActor.Chains, newChains))
			require.Equal(t, nActor.StakedTokens, newStakedTokens, "update stakedTokens")
		case "GetReadyToUnstake":
			readyToUnstake := false
			originalActor, err := getTestActor(db, actor.Address)
			require.NoError(t, err)
			if originalActor.UnstakingHeight == db.Height && originalActor.UnstakingHeight != DefaultUnstakingHeight {
				readyToUnstake = true
			}
			actors, err := db.ActorReadyToUnstakeWithChains(db.Height, readyToUnstakeQuery)
			require.NoError(t, err)
			if readyToUnstake {
				found := false
				for _, a := range actors {
					if originalActor.Address == hex.EncodeToString(a.Address) {
						found = true
						break
					}
				}
				if !found {
					fmt.Println(originalActor)
					fmt.Println(actors)
					fmt.Println(originalActor.UnstakingHeight, db.Height)
				}
				require.True(t, found, "readyToUnstake")
			} else {
				require.Nil(t, actors)
			}
		case "GetStatus":
			originalActor, err := getTestActor(db, actor.Address)
			require.NoError(t, err)
			addr, err := hex.DecodeString(originalActor.Address)
			require.NoError(t, err)
			status, err := db.GetActorStatus(addr, db.Height, actorStatusQuery)
			require.NoError(t, err)
			expectedStatus := 0
			switch {
			case originalActor.UnstakingHeight == DefaultUnstakingHeight:
				expectedStatus = persistence.StakedStatus
			case originalActor.UnstakingHeight > db.Height:
				expectedStatus = persistence.UnstakingStatus
			default:
				expectedStatus = persistence.UnstakedStatus
			}
			require.Equal(t, expectedStatus, status, "getStatus")
		case "GetPauseHeight":
			originalActor, err := getTestActor(db, actor.Address)
			require.NoError(t, err)
			addr, err := hex.DecodeString(originalActor.Address)
			require.NoError(t, err)
			pauseHeight, err := db.GetActorPauseHeightIfExists(addr, db.Height, actorPauseHeightIfExists)
			require.NoError(t, err)
			genericActor, err := db.GetActor(addr, db.Height, getActorQuery, getChainsQuery)
			require.Equal(t, int(originalActor.PausedHeight), int(pauseHeight), "getPauseHeight "+fmt.Sprintf("%d", genericActor.UnstakingHeight))
		case "SetUnstakingHeight":
			originalActor, err := getTestActor(db, actor.Address)
			require.NoError(t, err)
			newUnstakingHeight := rand.Int63()
			addr, err := hex.DecodeString(originalActor.Address)
			require.NoError(t, err)
			err = db.SetActorUnstakingHeightAndStatus(addr, newUnstakingHeight, setActorUnstakingHeight)
			require.NoError(t, err)
			nActor, err := getTestActor(db, originalActor.Address)
			require.NoError(t, err)
			require.Equal(t, int(newUnstakingHeight), int(nActor.UnstakingHeight), "setUnstakingHeight")
		case "SetPauseHeight":
			originalActor, err := getTestActor(db, actor.Address)
			require.NoError(t, err)
			newPauseHeight := rand.Int63()
			addr, err := hex.DecodeString(originalActor.Address)
			require.NoError(t, err)
			err = db.SetActorPauseHeight(addr, newPauseHeight, setActorPauseHeightQuery)
			require.NoError(t, err)
			nActor, err := getTestActor(db, actor.Address)
			require.NoError(t, err)
			require.Equal(t, int(newPauseHeight), int(nActor.PausedHeight), "setPauseHeight")
		case "SetPausedToUnstaking":
			randomUnstakingHeight := db.Height + int64(rand.Intn(15))
			isPausedAndReadyToUnstake := false
			originalActor, err := getTestActor(db, actor.Address)
			require.NoError(t, err)
			if originalActor.PausedHeight != DefaultPauseHeight && db.Height > originalActor.PausedHeight {
				isPausedAndReadyToUnstake = true
			}
			err = db.SetActorStatusAndUnstakingHeightPausedBefore(db.Height, randomUnstakingHeight, setActorUnstakingPausedBefore)
			require.NoError(t, err)
			nActor, err := getTestActor(db, originalActor.Address)
			if isPausedAndReadyToUnstake {
				require.Equal(t, int(nActor.UnstakingHeight), int(randomUnstakingHeight), "setPausedToUnstaking")
			}
		case "GetOutput":
			originalActor, err := getTestActor(db, actor.Address)
			require.NoError(t, err)
			addr, err := hex.DecodeString(originalActor.Address)
			require.NoError(t, err)
			outputAddr, err := db.GetActorOutputAddress(addr, db.Height, getActorOutputAddress)
			require.NoError(t, err)
			require.Equal(t, originalActor.OutputAddress, hex.EncodeToString(outputAddr), "getOutput")
		case "NextHeight":
			db.Height++
		}
	})
}

func GetRandomChains() (chains []string) {
	rand.Seed(time.Now().UnixNano())
	letterBytes := "ABCDEF0123456789"
	iterations := rand.Intn(14) + 1
	dupMap := make(map[string]struct{})
	for i := 0; i < iterations; i++ {
		b := make([]byte, 4)
		for i := range b {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
		if _, found := dupMap[string(b)]; found {
			i--
			continue
		}
		dupMap[string(b)] = struct{}{}
		chains = append(chains, string(b))
	}
	return
}

func isUnorderedEqual(slice1, slice2 []string) (isEqual bool) {
	if len(slice1) != len(slice2) {
		return false
	}
	compare := make(map[string]int)
	for _, s := range slice1 {
		compare[s]++
	}
	for _, s := range slice2 {
		if _, ok := compare[s]; !ok {
			return false
		}
		compare[s] = compare[s] - 1
		if compare[s] == 0 {
			delete(compare, s)
		}
	}
	return len(compare) == 0
}
