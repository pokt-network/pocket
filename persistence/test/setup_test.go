package test

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
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

	DefaultStakeStatus     = int32(persistence.StakedStatus)
	DefaultPauseHeight     = int64(-1)
	DefaultUnstakingHeight = int64(-1)

	OlshanskyURL    = "https://olshansky.info"
	OlshanskyChains = []string{"OLSH"}

	testSchema = "test_schema"
)
var testPersistenceMod modules.PersistenceModule // initialized in TestMain

// See https://github.com/ory/dockertest as reference for the template of this code
// Postgres example can be found here: https://github.com/ory/dockertest/blob/v3/examples/PostgreSQL.md
func TestMain(m *testing.M) {
	pool, resource, dbUrl := test_artifacts.SetupPostgresDocker()
	testPersistenceMod = newTestPersistenceModule(dbUrl)
	exitCode := m.Run()
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func NewTestPostgresContext(t *testing.T, height int64) *persistence.PostgresContext {
	ctx, err := testPersistenceMod.NewRWContext(height)
	require.NoError(t, err)

	db, ok := ctx.(*persistence.PostgresContext)
	require.True(t, ok)

	t.Cleanup(func() {
		require.NoError(t, db.Release())
		require.NoError(t, db.ResetContext())
	})

	return db
}

func NewFuzzTestPostgresContext(f *testing.F, height int64) *persistence.PostgresContext {
	ctx, err := testPersistenceMod.NewRWContext(height)
	if err != nil {
		log.Fatalf("Error creating new context: %v\n", err)
	}

	db, ok := ctx.(*persistence.PostgresContext)
	if !ok {
		log.Fatalf("Error casting RW context to Postgres context")
	}

	f.Cleanup(func() {
		if err := db.Release(); err != nil {
			f.FailNow()
		}
		if err := db.ResetContext(); err != nil {
			f.FailNow()
		}
	})

	return db
}

// TODO(andrew): Take in `t testing.T` as a parameter and error if there's an issue
func newTestPersistenceModule(databaseUrl string) modules.PersistenceModule {
	cfg := runtime.NewConfig(&runtime.BaseConfig{}, runtime.WithPersistenceConfig(&types.PersistenceConfig{
		PostgresUrl:    databaseUrl,
		NodeSchema:     testSchema,
		BlockStorePath: "",
	}))
	genesisState, _ := test_artifacts.NewGenesisState(5, 1, 1, 1)
	runtimeCfg := runtime.NewManager(cfg, genesisState)

	persistenceMod, err := persistence.Create(runtimeCfg)
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	return persistenceMod.(modules.PersistenceModule)
}

// IMPROVE(team): Extend this to more complex and variable test cases challenging & randomizing the state of persistence.
func fuzzSingleProtocolActor(
	f *testing.F,
	newTestActor func() (types.BaseActor, error),
	getTestActor func(db *persistence.PostgresContext, address string) (*types.BaseActor, error),
	protocolActorSchema types.ProtocolActorSchema) {

	db := NewFuzzTestPostgresContext(f, 0)

	err := db.DebugClearAll()
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
					case types.ServiceURLCol:
						newActorSpecificParam = getRandomServiceURL()
					case types.MaxRelaysCol:
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
			updatedActor := types.BaseActor{
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
			require.NotContains(t, newActor.StakedTokens, "invalid")
			// TODO(andrew): Use `require.Contains` instead. E.g. require.NotContains(t, newActor.StakedTokens, "invalid")
			if strings.Contains(newActor.StakedTokens, "invalid") {
				log.Println("")
			}
			require.Equal(t, newActor.StakedTokens, newStakedTokens, "staked tokens not updated")
			require.Equal(t, newActor.ActorSpecificParam, newActorSpecificParam, "actor specific param not updated")
		case "GetActorsReadyToUnstake":
			unstakingActors, err := db.GetActorsReadyToUnstake(protocolActorSchema, db.Height)
			require.NoError(t, err)

			if originalActor.UnstakingHeight != db.Height { // Not ready to unstake
				require.Nil(t, unstakingActors)
			} else {
				idx := slices.IndexFunc(unstakingActors, func(a modules.IUnstakingActor) bool {
					return originalActor.Address == hex.EncodeToString(a.GetAddress())
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
