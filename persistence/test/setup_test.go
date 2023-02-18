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
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/runtime/test_artifacts/keygenerator"
	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
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

	DefaultAccountAmount = converters.BigIntToString(DefaultAccountBig)
	DefaultStake         = converters.BigIntToString(DefaultStakeBig)
	DefaultMaxRelays     = converters.BigIntToString(DefaultMaxRelaysBig)
	StakeToUpdate        = converters.BigIntToString((&big.Int{}).Add(DefaultStakeBig, DefaultDeltaBig))

	DefaultStakeStatus     = int32(persistence.StakedStatus)
	DefaultPauseHeight     = int64(-1) // pauseHeight=-1 means not paused
	DefaultUnstakingHeight = int64(-1) // pauseHeight=-1 means not unstaking

	OlshanskyURL    = "https://olshansky.info"
	OlshanskyChains = []string{"OLSH"}

	testSchema = "test_schema"

	genesisStateNumValidators   = 5
	genesisStateNumServicers    = 1
	genesisStateNumApplications = 1
	genesisStateNumFishermen    = 1
)
var testPersistenceMod modules.PersistenceModule // initialized in TestMain

// See https://github.com/ory/dockertest as reference for the template of this code
// Postgres example can be found here: https://github.com/ory/dockertest/blob/v3/examples/PostgreSQL.md
func TestMain(m *testing.M) {
	pool, resource, dbUrl := test_artifacts.SetupPostgresDocker()
	testPersistenceMod = newTestPersistenceModule(dbUrl)
	if testPersistenceMod == nil {
		log.Fatal("[ERROR] Unable to create new test persistence module")
	}
	exitCode := m.Run()
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

// IMPROVE: Look into returning `testPersistenceMod` to avoid exposing underlying abstraction.
func NewTestPostgresContext(t testing.TB, height int64) *persistence.PostgresContext {
	ctx, err := testPersistenceMod.NewRWContext(height)
	if err != nil {
		log.Fatalf("Error creating new context: %v\n", err)
	}

	db, ok := ctx.(*persistence.PostgresContext)
	if !ok {
		log.Fatalf("Error casting RW context to Postgres context")
	}

	// TECHDEBT: This should not be part of `NewTestPostgresContext`. It causes unnecessary resets
	// if we call `NewTestPostgresContext` more than once in a single test.
	t.Cleanup(resetStateToGenesis)

	return db
}

// TODO(olshansky): Take in `t testing.T` as a parameter and error if there's an issue
func newTestPersistenceModule(databaseUrl string) modules.PersistenceModule {
	teardownDeterministicKeygen := keygenerator.GetInstance().SetSeed(42)
	defer teardownDeterministicKeygen()

	cfg := &configs.Config{
		Persistence: &configs.PersistenceConfig{
			PostgresUrl:       databaseUrl,
			NodeSchema:        testSchema,
			BlockStorePath:    "",
			TxIndexerPath:     "",
			TreesStoreDir:     "",
			MaxConnsCount:     4,
			MinConnsCount:     0,
			MaxConnLifetime:   "1h",
			MaxConnIdleTime:   "30m",
			HealthCheckPeriod: "5m",
		},
	}

	genesisState, _ := test_artifacts.NewGenesisState(
		genesisStateNumValidators,
		genesisStateNumServicers,
		genesisStateNumApplications,
		genesisStateNumServicers,
	)
	runtimeMgr := runtime.NewManager(cfg, genesisState)
	bus, err := runtime.CreateBus(runtimeMgr)
	if err != nil {
		log.Printf("Error creating bus: %s", err)
		return nil
	}

	persistenceMod, err := persistence.Create(bus)
	if err != nil {
		log.Printf("Error creating persistence module: %s", err)
		return nil
	}

	return persistenceMod.(modules.PersistenceModule)
}

// IMPROVE(team): Extend this to more complex and variable test cases challenging & randomizing the state of persistence.
//
//nolint:gosec // G404 - Weak random source is okay in unit tests
func fuzzSingleProtocolActor(
	f *testing.F,
	newTestActor func() (*coreTypes.Actor, error),
	getTestActor func(db *persistence.PostgresContext, address string) (*coreTypes.Actor, error),
	protocolActorSchema types.ProtocolActorSchema,
) {
	// Clear the genesis state.
	clearAllState()
	db := NewTestPostgresContext(f, 0)

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
			newStakedTokens := originalActor.StakedAmount
			newChains := originalActor.Chains
			newActorSpecificParam := originalActor.GenericParam

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
			updatedActor := &coreTypes.Actor{
				Address:         originalActor.Address,
				PublicKey:       originalActor.PublicKey,
				StakedAmount:    newStakedTokens,
				GenericParam:    newActorSpecificParam,
				Output:          originalActor.Output,
				PausedHeight:    originalActor.PausedHeight,
				UnstakingHeight: originalActor.UnstakingHeight,
				Chains:          newChains,
			}
			err = db.UpdateActor(protocolActorSchema, updatedActor)
			require.NoError(t, err)

			newActor, err := getTestActor(db, originalActor.Address)
			require.NoError(t, err)

			require.ElementsMatch(t, newActor.Chains, newChains, "staked chains not updated")
			require.NotContains(t, newActor.StakedAmount, "invalid")
			// TODO(andrew): Use `require.Contains` instead. E.g. require.NotContains(t, newActor.StakedTokens, "invalid")
			if strings.Contains(newActor.StakedAmount, "invalid") {
				log.Println("")
			}
			require.Equal(t, newStakedTokens, newActor.StakedAmount, "staked tokens not updated")
			require.Equal(t, newActorSpecificParam, newActor.GenericParam, "actor specific param not updated")
		case "GetActorsReadyToUnstake":
			unstakingActors, err := db.GetActorsReadyToUnstake(protocolActorSchema, db.Height)
			require.NoError(t, err)

			if originalActor.UnstakingHeight != db.Height { // Not ready to unstake
				require.Nil(t, unstakingActors)
			} else {
				idx := slices.IndexFunc(unstakingActors, func(a *moduleTypes.UnstakingActor) bool {
					return originalActor.Address == a.Address
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

			if db.Height > originalActor.PausedHeight && originalActor.PausedHeight != DefaultPauseHeight { // isPausedAndReadyToUnstake
				require.Equal(t, newUnstakingHeight, newActor.UnstakingHeight, "setPausedToUnstaking")
			}
		case "GetActorOutputAddr":
			outputAddr, err := db.GetActorOutputAddress(protocolActorSchema, addr, db.Height)
			require.NoError(t, err)

			require.Equal(t, originalActor.Output, hex.EncodeToString(outputAddr), "output address incorrect")
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
	//nolint:gosec // G404 - Weak random source is okay in unit tests
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

//nolint:gosec // G404 - Weak random source is okay in unit tests
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
	return converters.BigIntToString(big.NewInt(rand.Int63())) //nolint:gosec // G404 - Weak random source is okay in unit tests
}

func setRandomSeed() {
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck // G404 - Weak random source is okay in unit tests
}

// This is necessary for unit tests that are dependant on a baseline genesis state
func resetStateToGenesis() {
	if err := testPersistenceMod.ReleaseWriteContext(); err != nil {
		log.Fatalf("Error releasing write context: %v\n", err)
	}
	if err := testPersistenceMod.HandleDebugMessage(&messaging.DebugMessage{
		Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
		Message: nil,
	}); err != nil {
		log.Fatalf("Error clearing state: %v\n", err)
	}
}

// This is necessary for unit tests that are dependant on a completely clear state when starting
func clearAllState() {
	if err := testPersistenceMod.ReleaseWriteContext(); err != nil {
		log.Fatalf("Error releasing write context: %v\n", err)
	}
	if err := testPersistenceMod.HandleDebugMessage(&messaging.DebugMessage{
		Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_CLEAR_STATE,
		Message: nil,
	}); err != nil {
		log.Fatalf("Error clearing state: %v\n", err)
	}
}
