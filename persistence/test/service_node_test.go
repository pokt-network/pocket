package test

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func FuzzServiceNode(f *testing.F) {
	fuzzSingleProtocolActor(f,
		NewTestGenericActor(schema.ServiceNodeActor, newTestServiceNode),
		GetGenericActor(schema.ServiceNodeActor, getTestServiceNode),
		schema.ServiceNodeActor)
}

func TestInsertServiceNodeAndExists(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *testPostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	db.Height = 1

	serviceNode2, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	exists, err := db.GetServiceNodeExists(serviceNode.Address, 0)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at previous height does not")

	exists, err = db.GetServiceNodeExists(serviceNode.Address, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")

	exists, err = db.GetServiceNodeExists(serviceNode2.Address, 0)
	require.NoError(t, err)
	require.False(t, exists, "actor that should not exist at previous height appears to")

	exists, err = db.GetServiceNodeExists(serviceNode2.Address, 1)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist at current height does not")
}

func TestUpdateServiceNode(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *testPostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err := db.GetServiceNode(serviceNode.Address, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for current height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for current height")

	db.Height = 1

	require.NotEqual(t, DefaultStake, StakeToUpdate)   // sanity check to make sure the tests are correct
	require.NotEqual(t, DefaultChains, ChainsToUpdate) // sanity check to make sure the tests are correct
	err = db.UpdateServiceNode(serviceNode.Address, serviceNode.ServiceUrl, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetServiceNode(serviceNode.Address, 0)
	require.NoError(t, err)
	require.Equal(t, DefaultChains, chains, "default chains incorrect for previous height")
	require.Equal(t, DefaultStake, stakedTokens, "default stake incorrect for previous height")

	_, _, stakedTokens, _, _, _, _, chains, err = db.GetServiceNode(serviceNode.Address, 1)
	require.NoError(t, err)
	require.Equal(t, ChainsToUpdate, chains, "chains not updated for current height")
	require.Equal(t, StakeToUpdate, stakedTokens, "stake not updated for current height")
}

func TestGetServiceNodesReadyToUnstake(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *testPostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	serviceNode2, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	serviceNode3, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	// Unstake serviceNode at height 0
	err = db.SetServiceNodeUnstakingHeightAndStatus(serviceNode.Address, 0, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Unstake serviceNode2 and serviceNode3 at height 1
	err = db.SetServiceNodeUnstakingHeightAndStatus(serviceNode2.Address, 1, persistence.UnstakingStatus)
	require.NoError(t, err)
	err = db.SetServiceNodeUnstakingHeightAndStatus(serviceNode3.Address, 1, persistence.UnstakingStatus)
	require.NoError(t, err)

	// Check unstaking serviceNodes at height 0
	unstakingServiceNodes, err := db.GetServiceNodesReadyToUnstake(0, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 1, len(unstakingServiceNodes), "wrong number of actors ready to unstake at height 0")
	require.Equal(t, serviceNode.Address, unstakingServiceNodes[0].Address, "unexpected serviceNodelication actor returned")

	// Check unstaking serviceNodes at height 1
	unstakingServiceNodes, err = db.GetServiceNodesReadyToUnstake(1, persistence.UnstakingStatus)
	require.NoError(t, err)
	require.Equal(t, 2, len(unstakingServiceNodes), "wrong number of actors ready to unstake at height 1")
	require.ElementsMatch(t, [][]byte{serviceNode2.Address, serviceNode3.Address}, [][]byte{unstakingServiceNodes[0].Address, unstakingServiceNodes[1].Address})
}

func TestGetServiceNodeStatus(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 1, // intentionally set to a non-zero height
		DB:     *testPostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	// Check status before the serviceNode exists
	status, err := db.GetServiceNodeStatus(serviceNode.Address, 0)
	require.Error(t, err)
	require.Equal(t, status, persistence.UndefinedStakingStatus, "unexpected status")

	// Check status after the serviceNode exists
	status, err = db.GetServiceNodeStatus(serviceNode.Address, 1)
	require.NoError(t, err)
	require.Equal(t, status, DefaultStakeStatus, "unexpected status")
}

func TestGetServiceNodePauseHeightIfExists(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 1, // intentionally set to a non-zero height
		DB:     *testPostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	// Check pause height when serviceNode does not exist
	pauseHeight, err := db.GetServiceNodePauseHeightIfExists(serviceNode.Address, 0)
	require.Error(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")

	// Check pause height when serviceNode does not exist
	pauseHeight, err = db.GetServiceNodePauseHeightIfExists(serviceNode.Address, 1)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, DefaultPauseHeight, "unexpected pause height")
}

func TestSetServiceNodePauseHeightAndUnstakeLater(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *testPostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	pauseHeight := int64(1)
	unstakingHeight := pauseHeight + 10

	err = db.SetServiceNodePauseHeight(serviceNode.Address, pauseHeight)
	require.NoError(t, err)

	_, _, _, _, _, serviceNodePausedHeight, _, _, err := db.GetServiceNode(serviceNode.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, serviceNodePausedHeight, "pause height not updated")

	err = db.SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pauseHeight+1, unstakingHeight, -1 /*unused*/)
	require.NoError(t, err)

	_, _, _, _, _, _, serviceNodeUnstakingHeight, _, err := db.GetServiceNode(serviceNode.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, unstakingHeight, serviceNodeUnstakingHeight, "unstaking height was not set correctly")
}

func TestGetServiceNodeOutputAddress(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *testPostgresDB,
	}

	serviceNode, err := createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	output, err := db.GetServiceNodeOutputAddress(serviceNode.Address, 0)
	require.NoError(t, err)
	require.Equal(t, output, serviceNode.Output, "unexpected output address")
}

func TestGetAllServiceNodes(t *testing.T) {
	db := &persistence.PostgresContext{
		Height: 0,
		DB:     *testPostgresDB,
	}

	// The default test state contains 1 service node
	serviceNodes, err := db.GetAllServiceNodes(0)
	require.NoError(t, err)
	require.Len(t, serviceNodes, 1)

	// Add 2 services nodes at height 1
	db.Height++
	_, err = createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)
	_, err = createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	// 1 services nodes at height 0
	serviceNodes, err = db.GetAllServiceNodes(0)
	require.NoError(t, err)
	require.Len(t, serviceNodes, 1)

	// 3 services nodes at height 1
	serviceNodes, err = db.GetAllServiceNodes(1)
	require.NoError(t, err)
	require.Len(t, serviceNodes, 3)

	// Add 1 services nodes at height 3
	db.Height++
	db.Height++
	_, err = createAndInsertDefaultTestServiceNode(db)
	require.NoError(t, err)

	// 1 services nodes at height 0
	serviceNodes, err = db.GetAllServiceNodes(0)
	require.NoError(t, err)
	require.Len(t, serviceNodes, 1)

	// 3 services nodes at height 1
	serviceNodes, err = db.GetAllServiceNodes(1)
	require.NoError(t, err)
	require.Len(t, serviceNodes, 3)

	// 4 services nodes at height 2
	serviceNodes, err = db.GetAllServiceNodes(2)
	require.NoError(t, err)
	require.Len(t, serviceNodes, 3)

	// 4 services nodes at height 3
	serviceNodes, err = db.GetAllServiceNodes(3)
	require.NoError(t, err)
	require.Len(t, serviceNodes, 4)

	// Update the service nodes at different heights and confirm that count does not change
	for _, sn := range serviceNodes {
		db.Height++
		err = db.UpdateServiceNode(sn.Address, sn.ServiceUrl, sn.StakedTokens, []string{"ABBA"})
		require.NoError(t, err)

		// 4 service nodes at new height
		serviceNodes, err := db.GetAllServiceNodes(db.Height)
		require.NoError(t, err)
		require.Len(t, serviceNodes, 4)
	}

	// 3 services nodes at height 1
	serviceNodes, err = db.GetAllServiceNodes(1)
	require.NoError(t, err)
	require.Len(t, serviceNodes, 3)

	// 4 services nodes at height 10
	serviceNodes, err = db.GetAllServiceNodes(10)
	require.NoError(t, err)
	require.Len(t, serviceNodes, 4)

	// DISCUSS_IN_THIS_COMMIT: Since we do not support `DeleteActor`, should we filter here based on status? If so, tests need to be updated.
	for _, sn := range serviceNodes {
		db.Height++
		err = db.DeleteServiceNode(sn.Address)
		require.NoError(t, err)
	}
}

func newTestServiceNode() (*typesGenesis.ServiceNode, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &typesGenesis.ServiceNode{
		Address:         operatorKey.Address(),
		PublicKey:       operatorKey.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		ServiceUrl:      DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          outputAddr,
	}, nil
}

func createAndInsertDefaultTestServiceNode(db *persistence.PostgresContext) (*typesGenesis.ServiceNode, error) {
	serviceNode, err := newTestServiceNode()
	if err != nil {
		return nil, err
	}

	return serviceNode, db.InsertServiceNode(
		serviceNode.Address,
		serviceNode.PublicKey,
		serviceNode.Output,
		false,
		DefaultStakeStatus,
		DefaultServiceUrl,
		DefaultStake,
		DefaultChains,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
}

func getTestServiceNode(db persistence.PostgresContext, address []byte) (*typesGenesis.ServiceNode, error) {
	operator, publicKey, stakedTokens, serviceURL, outputAddress, pauseHeight, unstakingHeight, chains, err := db.GetServiceNode(address, db.Height)
	if err != nil {
		return nil, err
	}

	operatorAddr, err := hex.DecodeString(operator)
	if err != nil {
		return nil, err
	}

	operatorPubKey, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}

	outputAddr, err := hex.DecodeString(outputAddress)
	if err != nil {
		return nil, err
	}

	return &typesGenesis.ServiceNode{
		Address:         operatorAddr,
		PublicKey:       operatorPubKey,
		Paused:          false,
		Status:          persistence.UnstakingHeightToStatus(unstakingHeight),
		Chains:          chains,
		ServiceUrl:      serviceURL,
		StakedTokens:    stakedTokens,
		PausedHeight:    pauseHeight,
		UnstakingHeight: unstakingHeight,
		Output:          outputAddr,
	}, nil
}
