package test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"testing"

	typesPers "github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared/test_artifacts"
	utilTypes "github.com/pokt-network/pocket/utility/types"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility"
	"github.com/stretchr/testify/require"
)

var (
	defaultTestingChainsEdited = []string{"0002"}
	defaultUnstaking           = int64(2017)
	defaultSendAmount          = big.NewInt(10000)
	defaultNonceString         = utilTypes.BigIntToString(test_artifacts.DefaultAccountAmount)
	defaultSendAmountString    = utilTypes.BigIntToString(defaultSendAmount)
	testSchema                 = "test_schema"
)

var testPersistenceMod modules.PersistenceModule

func NewTestingMempool(_ *testing.T) utilTypes.Mempool {
	return utilTypes.NewMempool(1000000, 1000)
}

func TestMain(m *testing.M) {
	pool, resource, dbUrl := test_artifacts.SetupPostgresDocker()
	testPersistenceMod = newTestPersistenceModule(dbUrl)
	m.Run()
	os.Remove(testingConfigFilePath)
	os.Remove(testingGenesisFilePath)
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
}

func NewTestingUtilityContext(t *testing.T, height int64) utility.UtilityContext {
	persistenceContext, err := testPersistenceMod.NewRWContext(height)
	require.NoError(t, err)

	t.Cleanup(func() {
		testPersistenceMod.ResetContext()
	})

	return utility.UtilityContext{
		LatestHeight: height,
		Mempool:      NewTestingMempool(t),
		Context: &utility.Context{
			PersistenceRWContext: persistenceContext,
			SavePointsM:          make(map[string]struct{}),
			SavePoints:           make([][]byte, 0),
		},
	}
}

// TODO(andrew): Take in `t` and fail the test if there's an error
func newTestPersistenceModule(databaseUrl string) modules.PersistenceModule {
	cfg := modules.Config{
		Persistence: &typesPers.PersistenceConfig{
			PostgresUrl:    databaseUrl,
			NodeSchema:     testSchema,
			BlockStorePath: "",
		},
	}
	// TODO(andrew): Move the number of actors into local constants
	genesisState, _ := test_artifacts.NewGenesisState(5, 1, 1, 1)
	createTestingGenesisAndConfigFiles(cfg, genesisState)
	runtimeCfg := runtime.New(testingConfigFilePath, testingGenesisFilePath)

	persistenceMod, err := persistence.Create(runtimeCfg) // TODO (Drewsky) this is the last remaining cross module import and needs a fix...
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	persistenceMod.Start() // TODO: Check for error
	return persistenceMod.(modules.PersistenceModule)
}

const (
	testingGenesisFilePath = "genesis.json"
	testingConfigFilePath  = "config.json"
)

func createTestingGenesisAndConfigFiles(cfg modules.Config, genesisState modules.GenesisState) {
	config, err := json.Marshal(cfg.Persistence)
	if err != nil {
		log.Fatal(err)
	}
	genesis, err := json.Marshal(genesisState.PersistenceGenesisState)
	if err != nil {
		log.Fatal(err)
	}
	genesisFile := make(map[string]json.RawMessage)
	configFile := make(map[string]json.RawMessage)
	persistenceModuleName := persistence.PersistenceModuleName
	genesisFile[test_artifacts.GetGenesisFileName(persistenceModuleName)] = genesis
	configFile[persistenceModuleName] = config
	genesisFileBz, err := json.MarshalIndent(genesisFile, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	configFileBz, err := json.MarshalIndent(configFile, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(testingGenesisFilePath, genesisFileBz, 0777); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(testingConfigFilePath, configFileBz, 0777); err != nil {
		log.Fatal(err)
	}
}
