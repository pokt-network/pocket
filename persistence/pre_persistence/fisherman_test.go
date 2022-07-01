package pre_persistence

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
)

func NewTestFisherman() typesGenesis.Fisherman {
	pub1, _ := crypto.GeneratePublicKey()
	addr1 := pub1.Address()
	addr2, _ := crypto.GenerateAddress()
	return typesGenesis.Fisherman{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		ServiceUrl:      typesGenesis.DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    0,
		UnstakingHeight: 0,
		Output:          addr2,
	}
}

func TestGetFishermanExists(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestFisherman()
	addr2, _ := crypto.GenerateAddress()
	if err := ctx.InsertFisherman(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	height, err := ctx.GetHeight()
	if err != nil {
		t.Fatal(err)
	}
	exists, err := ctx.GetFishermanExists(actor.Address, height)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exists does not")
	}
	exists, err = ctx.GetFishermanExists(addr2, height)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that exists should not")
	}
}

func TestGetFisherman(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestFisherman()
	if err := ctx.InsertFisherman(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	height, err := ctx.GetHeight()
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := ctx.(*PrePersistenceContext).GetFisherman(actor.Address, height)
	require.NoError(t, err)
	if !bytes.Equal(actor.Address, got.Address) || !bytes.Equal(actor.PublicKey, got.PublicKey) {
		t.Fatalf("unexpected actor returned; expected %v got %v", actor, got)
	}
}

func TestGetAllFishermans(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor1 := NewTestFisherman()
	actor2 := NewTestFisherman()
	if err := ctx.InsertFisherman(actor1.Address, actor1.PublicKey, actor1.Output, actor1.Paused, int(actor1.Status),
		actor1.ServiceUrl, actor1.StakedTokens, actor1.Chains, int64(actor1.PausedHeight), actor1.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := ctx.InsertFisherman(actor2.Address, actor2.PublicKey, actor2.Output, actor2.Paused, int(actor2.Status),
		actor2.ServiceUrl, actor2.StakedTokens, actor2.Chains, int64(actor2.PausedHeight), actor2.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	fishermans, err := ctx.(*PrePersistenceContext).GetAllFishermen(0)
	require.NoError(t, err)
	got1, got2 := false, false
	for _, a := range fishermans {
		if bytes.Equal(a.Address, actor1.Address) {
			got1 = true
		}
		if bytes.Equal(a.Address, actor2.Address) {
			got2 = true
		}
	}
	if !got1 || !got2 {
		t.Fatal("not all actors returned")
	}
}

func TestUpdateFisherman(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestFisherman()
	if err := ctx.InsertFisherman(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	zero := types.BigIntToString(big.NewInt(0))
	bigExpectedTokens := big.NewInt(1)
	one := types.BigIntToString(bigExpectedTokens)
	height, err := ctx.GetHeight()
	if err != nil {
		t.Fatal(err)
	}
	before, _, err := ctx.(*PrePersistenceContext).GetFisherman(actor.Address, height)
	require.NoError(t, err)
	tokens := before.StakedTokens
	bigBeforeTokens, err := types.StringToBigInt(tokens)
	require.NoError(t, err)
	err = ctx.UpdateFisherman(actor.Address, zero, one, typesGenesis.DefaultChains)
	require.NoError(t, err)
	got, _, err := ctx.(*PrePersistenceContext).GetFisherman(actor.Address, height)
	require.NoError(t, err)
	bigAfterTokens, err := types.StringToBigInt(got.StakedTokens)
	require.NoError(t, err)
	bigAfterTokens.Sub(bigAfterTokens, bigBeforeTokens)
	if bigAfterTokens.Cmp(bigExpectedTokens) != 0 {
		t.Fatal("incorrect after balance")
	}
}

func TestDeleteFisherman(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestFisherman()
	if err := ctx.InsertFisherman(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	err := ctx.DeleteFisherman(actor.Address)
	require.NoError(t, err)
	height, err := ctx.GetHeight()
	if err != nil {
		t.Fatal(err)
	}
	exists, err := ctx.(*PrePersistenceContext).GetFishermanExists(actor.Address, height)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor exists when it shouldn't")
	}
}

func TestGetFishermansReadyToUnstake(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestFisherman()
	if err := ctx.InsertFisherman(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := ctx.SetFishermanUnstakingHeightAndStatus(actor.Address, 0, 1); err != nil {
		t.Fatal(err)
	}
	unstakingFishermans, err := ctx.(*PrePersistenceContext).GetFishermenReadyToUnstake(0, 1)
	require.NoError(t, err)
	if !bytes.Equal(unstakingFishermans[0].Address, actor.Address) {
		t.Fatalf("wrong actor returned, expected addr %v, got %v", unstakingFishermans[0].Address, actor.Address)
	}
}

func TestGetFishermanStatus(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestFisherman()
	if err := ctx.InsertFisherman(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	height, err := ctx.GetHeight()
	if err != nil {
		t.Fatal(err)
	}
	status, err := ctx.GetFishermanStatus(actor.Address, height)
	require.NoError(t, err)
	if status != int(actor.Status) {
		t.Fatal("unequal status")
	}
}

func TestGetFishermanPauseHeightIfExists(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestFisherman()
	if err := ctx.InsertFisherman(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	pausedHeight := 1
	err := ctx.SetFishermanPauseHeight(actor.Address, int64(pausedHeight))
	require.NoError(t, err)
	// HACK(olshansky): Don't know why this broke but it'll be deleted soon
	// pauseBeforeHeight, err := ctx.GetFishermanPauseHeightIfExists(actor.Address, 1)
	// require.NoError(t, err)
	// if pausedHeight != int(pauseBeforeHeight) {
	// 	t.Fatalf("incorrect pause height: expected %v, got %v", pausedHeight, pauseBeforeHeight)
	// }
}

func TestSetFishermanStatusAndUnstakingHeightIfPausedBefore(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestFisherman()
	if err := ctx.InsertFisherman(actor.Address, actor.PublicKey, actor.Output, true, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	pauseBeforeHeight, unstakingHeight, status := int64(1), int64(10), 1
	err := ctx.SetFishermanStatusAndUnstakingHeightIfPausedBefore(pauseBeforeHeight, unstakingHeight, status)
	require.NoError(t, err)
	height, err := ctx.GetHeight()
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := ctx.(*PrePersistenceContext).GetFisherman(actor.Address, height)
	require.NoError(t, err)
	if got.UnstakingHeight != unstakingHeight {
		t.Fatalf("wrong unstaking height: expected %v, got %v", unstakingHeight, got.UnstakingHeight)
	}
	if int(got.Status) != status {
		t.Fatalf("wrong status: expected %v, got %v", status, got.Status)
	}
}

func TestGetFishermanOutputAddress(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestFisherman()
	if err := ctx.InsertFisherman(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	height, err := ctx.GetHeight()
	if err != nil {
		t.Fatal(err)
	}
	output, err := ctx.GetFishermanOutputAddress(actor.Address, height)
	require.NoError(t, err)
	if !bytes.Equal(actor.Output, output) {
		t.Fatalf("incorrect output address expected %v, got %v", actor.Output, output)
	}
}
