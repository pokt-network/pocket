package test

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/require"
)

func TestInitParams(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}
	err := db.InitParams()
	require.NoError(t, err)
}

func TestGetSetIntParam(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitParams()
	require.NoError(t, err)

	newMaxChains := 42

	err = db.SetParam(types.AppMaxChainsParamName, newMaxChains)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	maxChains, err := db.GetIntParam(types.AppMaxChainsParamName, height)
	require.NoError(t, err)

	require.Equal(t, newMaxChains, maxChains)
}

func TestGetSetStringParam(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitParams()
	require.NoError(t, err)

	newServiceNodeMinimumStake := "99999999"

	err = db.SetParam(types.ServiceNodeMinimumStakeParamName, newServiceNodeMinimumStake)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	serviceNodeMinimumStake, err := db.GetStringParam(types.ServiceNodeMinimumStakeParamName, height)
	require.NoError(t, err)

	require.Equal(t, newServiceNodeMinimumStake, serviceNodeMinimumStake)
}

func TestGetSetByteArrayParam(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitParams()
	require.NoError(t, err)

	newOwner, err := hex.DecodeString("63585955783252764a6e576a5631647542486168426c63774e4655345a57617468545532637a6330516e4d5978575977674553537857644e4a6b4c7734575335416a65616c6d57494a47535364555933686d565a706e57564a6d6143526c54594248626864465a72646c624f646c59704a45536a6c6c52794d32527849545733566c6557464763745a465377466a57324a316157314562554a6c564b6c325470394753696c58544846474e786331567a70554d534a6c5335566d4c356f305157684663726c6b4e4a4e305931496c624a4e58537035554d4a7058564a705561506c3259484a47614b6c585a")
	require.NoError(t, err)

	err = db.SetParam(types.ServiceNodeUnstakingBlocksOwner, newOwner)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	owner, err := db.GetBytesParam(types.ServiceNodeUnstakingBlocksOwner, height)
	require.NoError(t, err)

	require.Equal(t, newOwner, owner)
}

func TestGetSetToggleIntFlag(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitFlags()
	require.NoError(t, err)

	newMaxChains := 42

	// insert with false
	err = db.SetFlag(types.AppMaxChainsParamName, newMaxChains, false)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	maxChains, enabled, err := db.GetIntFlag(types.AppMaxChainsParamName, height)
	require.NoError(t, err)

	require.Equal(t, newMaxChains, maxChains)

	require.Equal(t, false, enabled)

	// toggle to true
	err = db.SetFlag(types.AppMaxChainsParamName, newMaxChains, true)
	require.NoError(t, err)

	height, err = db.GetHeight()
	require.NoError(t, err)

	maxChains, enabled, err = db.GetIntFlag(types.AppMaxChainsParamName, height)
	require.NoError(t, err)

	require.Equal(t, newMaxChains, maxChains)

	require.Equal(t, true, enabled)
}

func TestGetSetToggleStringFlag(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitParams()
	require.NoError(t, err)

	newServiceNodeMinimumStake := "99999999"

	// insert with false
	err = db.SetFlag(types.ServiceNodeMinimumStakeParamName, newServiceNodeMinimumStake, false)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	serviceNodeMinimumStake, enabled, err := db.GetStringFlag(types.ServiceNodeMinimumStakeParamName, height)
	require.NoError(t, err)

	require.Equal(t, newServiceNodeMinimumStake, serviceNodeMinimumStake)
	require.Equal(t, false, enabled)

	//toggle to true
	err = db.SetFlag(types.ServiceNodeMinimumStakeParamName, newServiceNodeMinimumStake, true)
	require.NoError(t, err)

	height, err = db.GetHeight()
	require.NoError(t, err)

	serviceNodeMinimumStake, enabled, err = db.GetStringFlag(types.ServiceNodeMinimumStakeParamName, height)
	require.NoError(t, err)

	require.Equal(t, newServiceNodeMinimumStake, serviceNodeMinimumStake)
	require.Equal(t, true, enabled)

}

func TestGetSetToggleByteArrayFlag(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitParams()
	require.NoError(t, err)

	newOwner, err := hex.DecodeString("576c687353324648536b685a4d6d785159565677536c5a596345704e565456775531684f536d4a735354465a4d45354b546d7473636d4e47614664524d473831544731574e564e58624642685658417a5757704b4d4531466548524f56336872553064534d6c6b794d587068565868455532784b55303156634870574d574e34546b64475346525962476c54527a6c7756444a7353315274546e42526132524b5530645362316b7a62454e694d58425a55323134536d4a71515856564d573831576c525753466c555a456c5a656c5a35557a4e4f56314a7352545657566b303055573157616d52704d545a5256557033576a4677556d5274576c6c57526d6779546c684b546c5a52")
	require.NoError(t, err)

	// insert with false
	err = db.SetFlag(types.ServiceNodeUnstakingBlocksOwner, newOwner, false)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	owner, enabled, err := db.GetBytesFlag(types.ServiceNodeUnstakingBlocksOwner, height)
	require.NoError(t, err)

	require.Equal(t, newOwner, owner)
	require.Equal(t, false, enabled)

	//toggle to true
	err = db.SetFlag(types.ServiceNodeUnstakingBlocksOwner, newOwner, true)
	require.NoError(t, err)

	height, err = db.GetHeight()
	require.NoError(t, err)

	owner, enabled, err = db.GetBytesFlag(types.ServiceNodeUnstakingBlocksOwner, height)
	require.NoError(t, err)

	require.Equal(t, newOwner, owner)
	require.Equal(t, true, enabled)

}
