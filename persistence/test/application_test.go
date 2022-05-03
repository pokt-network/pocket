package test

import (
	"bytes"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"math/big"
	"testing"
)

var (
	DefaultChains          = []string{"0001"}
	ChainsToUpdate         = []string{"0002"}
	DefaultServiceUrl      = "https://foo.bar"
	DefaultPoolName        = "TESTING_POOL"
	DefaultStakeBig        = big.NewInt(1000000000000000)
	DefaultStake           = types.BigIntToString(DefaultStakeBig)
	StakeToUpdate          = types.BigIntToString((&big.Int{}).Add(DefaultStakeBig, big.NewInt(100)))
	ParamToUpdate          = 2
	DefaultAccountBalance  = DefaultStake
	DefaultStakeStatus     = 2
	DefaultPauseHeight     = int64(-1)
	DefaultUnstakingHeight = int64(-1)
	PauseHeightToSet       = 1
)

func TestInsertAppAndExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	app2 := NewTestApp()
	if err := db.InsertApplication(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	exists, err := db.GetAppExists(app.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = db.GetAppExists(app2.Address)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("actor that should not exist, appears to")
	}
}

func TestUpdateApplication(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	if err := db.InsertApplication(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	_, _, stakedTokens, _, _, _, _, _, chains, err := db.GetApp(app.Address)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.UpdateApplication(app.Address, app.MaxRelays, StakeToUpdate, ChainsToUpdate); err != nil {
		t.Fatal(err)
	}
	_, _, stakedTokens, _, _, _, _, _, chains, err = db.GetApp(app.Address)
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

func TestDeleteApplication(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	if err := db.InsertApplication(app.Address, app.PublicKey, app.Output, false, DefaultStakeStatus, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, _, _, chains, err := db.GetApp(app.Address)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.DeleteApplication(app.Address); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, _, _, chains, err = db.GetApp(app.Address)
	if err != nil {
		t.Fatal(err)
	}
	if len(chains) != 0 {
		t.Fatal("chains not nullified")
	}
}

func TestGetAppsReadyToUnstake(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	if err := db.InsertApplication(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	// test SetAppUnstakingHeightAndStatus
	if err := db.SetAppUnstakingHeightAndStatus(app.Address, 0, 1); err != nil {
		t.Fatal(err)
	}
	// test GetAppsReadyToUnstake
	apps, err := db.GetAppsReadyToUnstake(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 1 {
		t.Fatal("wrong number of actors")
	}
	if !bytes.Equal(app.Address, apps[0].Address) {
		t.Fatal("unexpected actor returned")
	}
}

func TestGetAppStatus(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	if err := db.InsertApplication(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	status, err := db.GetAppStatus(app.Address)
	if err != nil {
		t.Fatal(err)
	}
	if status != DefaultStakeStatus {
		t.Fatalf("unexpected status: got %d expected %d", status, DefaultStakeStatus)
	}
}

func TestGetPauseHeightIfExists(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	if err := db.InsertApplication(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	height, err := db.GetAppPauseHeightIfExists(app.Address)
	if err != nil {
		t.Fatal(err)
	}
	if height != DefaultPauseHeight {
		t.Fatalf("unexpected pauseHeight: got %d expected %d", DefaultPauseHeight, DefaultStakeStatus)
	}
}

func TestSetAppsStatusAndUnstakingHeightPausedBefore(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	if err := db.InsertApplication(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := db.SetAppsStatusAndUnstakingHeightPausedBefore(1, 0, 1); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, unstakingHeight, _, _, err := db.GetApp(app.Address)
	if err != nil {
		t.Fatal(err)
	}
	if unstakingHeight != 0 {
		t.Fatal("unexpected unstaking height")
	}
}

func TestSetAppPauseHeight(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	if err := db.InsertApplication(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, DefaultPauseHeight, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := db.SetAppPauseHeight(app.Address, int64(PauseHeightToSet)); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, pauseHeight, _, _, _, err := db.GetApp(app.Address)
	if err != nil {
		t.Fatal(err)
	}
	if pauseHeight != int64(PauseHeightToSet) {
		t.Fatal("unexpected pause height")
	}
}

func TestGetAppOutputAddress(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	app := NewTestApp()
	if err := db.InsertApplication(app.Address, app.PublicKey, app.Output, false, 1, DefaultStake, DefaultStake, DefaultChains, 0, DefaultUnstakingHeight); err != nil {
		t.Fatal(err)
	}
	output, err := db.GetAppOutputAddress(app.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output, app.Output) {
		t.Fatal("unexpected output address")
	}
}

func NewTestApp() typesGenesis.App {
	pub1, _ := crypto.GeneratePublicKey()
	addr1 := pub1.Address()
	addr2, _ := crypto.GenerateAddress()
	defaultMaxRelays := types.BigIntToString(big.NewInt(1000000))
	return typesGenesis.App{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		MaxRelays:       defaultMaxRelays,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    0,
		UnstakingHeight: 0,
		Output:          addr2,
	}
}
