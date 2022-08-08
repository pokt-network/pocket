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

	newOwner, _ := hex.DecodeString("Vh2MkBXWUhjeolEUDh2TNd2aTNFUlV0NOljRWR2Mud2dvdzVwkDROpGUQ5SOKNFU5MWbZJDeyI1aod0YZZkMjdUMyk1dVpnVhBXbWhEcxolax0WW3JESXhmTuJFdOdUTxMGbXFnWxIVYkJjYFplRUVTVVZVavpWS1VTRXlmS5VmL5o0QWhFcrlkNJN0Y1IlbJNXSp5UMJpXVJpUaPl2YHJGaKlXZ")

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

	newOwner, _ := hex.DecodeString("WlhsS2FHSkhZMmxQYVVwSlZYcEpNVTVwU1hOSmJsSTFZME5KTmtscmNGaFdRMG81TG1WNVNXbFBhVXAzWWpKME1FeHROV3hrU0dSMlkyMXphVXhEU2xKU01VcHpWMWN4TkdGSFRYbGlTRzlwVDJsS1RtTnBRa2RKU0dSb1kzbENiMXBZU214SmJqQXVVMW81WlRWSFlUZElZelZ5UzNOV1JsRTVWVk00UW1WamRpMTZRVUp3WjFwUmRtWllWRmgyTlhKTlZR")

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
