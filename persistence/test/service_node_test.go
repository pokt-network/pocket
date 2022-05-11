package test

import (
	"bytes"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func TestInsertServiceNodeAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode()
	serviceNode2 := NewTestServiceNode()
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	exists, err := db.GetServiceNodeExists(serviceNode.Address)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = db.GetServiceNodeExists(serviceNode2.Address)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that should not exist, appears to")
	}
}

func TestUpdateServiceNode(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode()
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, _, chains, err := db.GetServiceNode(serviceNode.Address)
	require.NoError(t, err)
	err = db.UpdateServiceNode(serviceNode.Address, serviceNode.ServiceUrl, StakeToUpdate, ChainsToUpdate)
	require.NoError(t, err)
	_, _, stakedTokens, _, _, _, _, _, chains, err = db.GetServiceNode(serviceNode.Address)
	require.NoError(t, err)
	if chains[0] != ChainsToUpdate[0] {
		t.Fatal("chains not updated")
	}
	if stakedTokens != StakeToUpdate {
		t.Fatal("stake not updated")
	}
}

func TestDeleteServiceNode(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode()
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	_, _, _, _, _, _, _, _, chains, err := db.GetServiceNode(serviceNode.Address)
	require.NoError(t, err)
	err = db.DeleteServiceNode(serviceNode.Address)
	require.NoError(t, err)
	_, _, _, _, _, _, _, _, chains, err = db.GetServiceNode(serviceNode.Address)
	require.NoError(t, err)
	if len(chains) != 0 {
		t.Fatal("chains not nullified")
	}
}

func TestGetServiceNodesReadyToUnstake(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode()
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	// test SetServiceNodeUnstakingHeightAndStatus
	err = db.SetServiceNodeUnstakingHeightAndStatus(serviceNode.Address, 0, 1)
	require.NoError(t, err)
	// test GetServiceNodesReadyToUnstake
	serviceNodes, err := db.GetServiceNodesReadyToUnstake(0, 1)
	require.NoError(t, err)
	if len(serviceNodes) != 1 {
		t.Fatal("wrong number of actors")
	}
	if !bytes.Equal(serviceNode.Address, serviceNodes[0].Address) {
		t.Fatal("unexpected actor returned")
	}
}

func TestGetServiceNodeStatus(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode()
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	status, err := db.GetServiceNodeStatus(serviceNode.Address)
	require.NoError(t, err)
	if status != DefaultStakeStatus {
		t.Fatalf("unexpected status: got %d expected %d", status, DefaultStakeStatus)
	}
}

func TestGetServiceNodePauseHeightIfExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode()
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight)
	require.NoError(t, err)
	height, err := db.GetServiceNodePauseHeightIfExists(serviceNode.Address)
	require.NoError(t, err)
	if height != DefaultPauseHeight {
		t.Fatalf("unexpected pauseHeight: got %d expected %d", DefaultPauseHeight, DefaultStakeStatus)
	}
}

func TestSetServiceNodesStatusAndUnstakingHeightPausedBefore(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode()
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetServiceNodesStatusAndUnstakingHeightPausedBefore(1, 0, 1)
	require.NoError(t, err)
	_, _, _, _, _, _, unstakingHeight, _, _, err := db.GetServiceNode(serviceNode.Address)
	require.NoError(t, err)
	if unstakingHeight != 0 {
		t.Fatal("unexpected unstaking height")
	}
}

func TestSetServiceNodePauseHeight(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode()
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	err = db.SetServiceNodePauseHeight(serviceNode.Address, int64(PauseHeightToSet))
	require.NoError(t, err)
	_, _, _, _, _, pauseHeight, _, _, _, err := db.GetServiceNode(serviceNode.Address)
	require.NoError(t, err)
	if pauseHeight != int64(PauseHeightToSet) {
		t.Fatal("unexpected pause height")
	}
}

func TestGetServiceNodeOutputAddress(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode()
	err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight)
	require.NoError(t, err)
	output, err := db.GetServiceNodeOutputAddress(serviceNode.Address)
	require.NoError(t, err)
	if !bytes.Equal(output, serviceNode.Output) {
		t.Fatal("unexpected output address")
	}
}

func TestServiceNodeCount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	err := db.ClearAllDebug()
	require.NoError(t, err)
	count, err := db.GetServiceNodeCount(DefaultChains[0], 0)
	require.NoError(t, err)
	if count != 0 {
		t.Fatal("unexpected service node count")
	}
	serviceNode := NewTestServiceNode()
	err = db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, -1, DefaultUnstakingHeight)
	require.NoError(t, err)
	count, err = db.GetServiceNodeCount(DefaultChains[0], 0)
	require.NoError(t, err)
	if count != 1 {
		t.Fatal("unexpected service node count")
	}
}

func NewTestServiceNode() typesGenesis.ServiceNode {
	pub1, _ := crypto.GeneratePublicKey()
	addr1 := pub1.Address()
	addr2, _ := crypto.GenerateAddress()
	return typesGenesis.ServiceNode{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		ServiceUrl:      DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    0,
		UnstakingHeight: 0,
		Output:          addr2,
	}
}
