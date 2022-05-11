package test

import (
	"bytes"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
)

func TestInsertServiceNodeAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	serviceNode := NewTestServiceNode()
	serviceNode2 := NewTestServiceNode()
	if err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	exists, err := db.GetServiceNodeExists(serviceNode.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = db.GetServiceNodeExists(serviceNode2.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	_, _, stakedTokens, _, _, _, _, _, chains, err := db.GetServiceNode(serviceNode.Address)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.UpdateServiceNode(serviceNode.Address, serviceNode.ServiceUrl, StakeToUpdate, ChainsToUpdate); err != nil {
		t.Fatal(err)
	}
	_, _, stakedTokens, _, _, _, _, _, chains, err = db.GetServiceNode(serviceNode.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, _, _, chains, err := db.GetServiceNode(serviceNode.Address)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.DeleteServiceNode(serviceNode.Address); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, _, _, chains, err = db.GetServiceNode(serviceNode.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	// test SetServiceNodeUnstakingHeightAndStatus
	if err := db.SetServiceNodeUnstakingHeightAndStatus(serviceNode.Address, 0, 1); err != nil {
		t.Fatal(err)
	}
	// test GetServiceNodesReadyToUnstake
	serviceNodes, err := db.GetServiceNodesReadyToUnstake(0, 1)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	status, err := db.GetServiceNodeStatus(serviceNode.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	height, err := db.GetServiceNodePauseHeightIfExists(serviceNode.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := db.SetServiceNodesStatusAndUnstakingHeightPausedBefore(1, 0, 1); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, unstakingHeight, _, _, err := db.GetServiceNode(serviceNode.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := db.SetServiceNodePauseHeight(serviceNode.Address, int64(PauseHeightToSet)); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, pauseHeight, _, _, _, err := db.GetServiceNode(serviceNode.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	output, err := db.GetServiceNodeOutputAddress(serviceNode.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output, serviceNode.Output) {
		t.Fatal("unexpected output address")
	}
}

func TestServiceNodeCount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	if err := db.ClearAllDebug(); err != nil {
		t.Fatal(err)
	}
	count, err := db.GetServiceNodeCount(DefaultChains[0], 0)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("unexpected service node count")
	}
	serviceNode := NewTestServiceNode()
	if err := db.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, -1, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	count, err = db.GetServiceNodeCount(DefaultChains[0], 0)
	if err != nil {
		t.Fatal(err)
	}
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
