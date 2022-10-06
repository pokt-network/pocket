package test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/pokt-network/pocket/shared/test_artifacts"
	utilTypes "github.com/pokt-network/pocket/utility/types"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility"
	"github.com/stretchr/testify/require"
)

const (
	testingValidatorCount   = 5
	testingServiceNodeCount = 1
	testingApplicationCount = 1
	testingFishermenCount   = 1
)

var (
	defaultTestingChainsEdited = []string{"0002"}
	defaultUnstaking           = int64(2017)
	defaultSendAmount          = big.NewInt(10000)
	defaultNonceString         = utilTypes.BigIntToString(test_artifacts.DefaultAccountAmount)
	defaultSendAmountString    = utilTypes.BigIntToString(defaultSendAmount)
	testSchema                 = "test_schema"
)

var actorTypes = []utilTypes.ActorType{
	utilTypes.ActorType_App,
	utilTypes.ActorType_ServiceNode,
	utilTypes.ActorType_Fisherman,
	utilTypes.ActorType_Validator,
}

var testPersistenceMod modules.PersistenceModule

func NewTestingMempool(_ *testing.T) utilTypes.Mempool {
	return utilTypes.NewMempool(1000000, 1000)
}

func TestMain(m *testing.M) {
	pool, resource, dbUrl := test_artifacts.SetupPostgresDocker()
	testPersistenceMod = newTestPersistenceModule(m, dbUrl)
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

func newTestPersistenceModule(_ *testing.M, databaseUrl string) modules.PersistenceModule {
	cfg := modules.Config{
		Persistence: &test_artifacts.MockPersistenceConfig{
			PostgresUrl:    databaseUrl,
			NodeSchema:     testSchema,
			BlockStorePath: "",
		},
	}

	genesisState, _ := test_artifacts.NewGenesisState(testingValidatorCount, testingServiceNodeCount, testingApplicationCount, testingFishermenCount)
	createTestingGenesisAndConfigFiles(cfg, genesisState)
	persistenceMod, err := persistence.Create(testingConfigFilePath, testingGenesisFilePath) // TODO (Olshansk) this is the last remaining cross module import and needs a fix...
	if err != nil {
		log.Fatal(err)
	}
	if err = persistenceMod.Start(); err != nil {
		log.Fatal(err)
	}
	return persistenceMod
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
	persistenceModuleName := new(persistence.PersistenceModule).GetModuleName()
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
