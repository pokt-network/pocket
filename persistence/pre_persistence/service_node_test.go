package pre_persistence

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
)

func NewTestServiceNode() ServiceNode {
	pub1, _ := crypto.GeneratePublicKey()
	addr1 := pub1.Address()
	addr2, _ := crypto.GenerateAddress()
	return ServiceNode{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          defaultStakeStatus,
		Chains:          defaultChains,
		ServiceUrl:      defaultServiceUrl,
		StakedTokens:    defaultStake,
		PausedHeight:    0,
		UnstakingHeight: 0,
		Output:          addr2,
	}
}

func TestGetServiceNodeExists(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestServiceNode()
	addr2, _ := crypto.GenerateAddress()
	if err := ctx.InsertServiceNode(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	exists, err := ctx.GetServiceNodeExists(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("actor that should exists does not")
	}
	exists, err = ctx.GetServiceNodeExists(addr2)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("actor that exists should not")
	}
}

func TestGetServiceNode(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestServiceNode()
	if err := ctx.InsertServiceNode(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	got, _, err := ctx.(*PrePersistenceContext).GetServiceNode(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actor.Address, got.Address) || !bytes.Equal(actor.PublicKey, got.PublicKey) {
		t.Fatalf("unexpected actor returned; expected %v got %v", actor, got)
	}
}

func TestGetAllServiceNodes(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor1 := NewTestServiceNode()
	actor2 := NewTestServiceNode()
	if err := ctx.InsertServiceNode(actor1.Address, actor1.PublicKey, actor1.Output, actor1.Paused, int(actor1.Status),
		actor1.ServiceUrl, actor1.StakedTokens, actor1.Chains, int64(actor1.PausedHeight), actor1.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := ctx.InsertServiceNode(actor2.Address, actor2.PublicKey, actor2.Output, actor2.Paused, int(actor2.Status),
		actor2.ServiceUrl, actor2.StakedTokens, actor2.Chains, int64(actor2.PausedHeight), actor2.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	serviceNodes, err := ctx.(*PrePersistenceContext).GetAllServiceNodes(0)
	if err != nil {
		t.Fatal(err)
	}
	got1, got2 := false, false
	for _, a := range serviceNodes {
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

func TestUpdateServiceNode(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestServiceNode()
	if err := ctx.InsertServiceNode(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	zero := BigIntToString(big.NewInt(0))
	bigExpectedTokens := big.NewInt(1)
	one := BigIntToString(bigExpectedTokens)
	before, _, err := ctx.(*PrePersistenceContext).GetServiceNode(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	tokens := before.StakedTokens
	bigBeforeTokens, err := StringToBigInt(tokens)
	if err != nil {
		t.Fatal(err)
	}
	err = ctx.UpdateServiceNode(actor.Address, zero, one, defaultChains)
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := ctx.(*PrePersistenceContext).GetServiceNode(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	bigAfterTokens, err := StringToBigInt(got.StakedTokens)
	if err != nil {
		t.Fatal(err)
	}
	bigAfterTokens.Sub(bigAfterTokens, bigBeforeTokens)
	if bigAfterTokens.Cmp(bigExpectedTokens) != 0 {
		t.Fatal("incorrect after balance")
	}
}

func TestDeleteServiceNode(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestServiceNode()
	if err := ctx.InsertServiceNode(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	err := ctx.DeleteServiceNode(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	exists, err := ctx.(*PrePersistenceContext).GetServiceNodeExists(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("actor exists when it shouldn't")
	}
}

func TestGetServiceNodesReadyToUnstake(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestServiceNode()
	if err := ctx.InsertServiceNode(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := ctx.SetServiceNodeUnstakingHeightAndStatus(actor.Address, 0, 1); err != nil {
		t.Fatal(err)
	}
	unstakingServiceNodes, err := ctx.(*PrePersistenceContext).GetServiceNodesReadyToUnstake(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(unstakingServiceNodes[0].Address, actor.Address) {
		t.Fatalf("wrong actor returned, expected addr %v, got %v", unstakingServiceNodes[0].Address, actor.Address)
	}
}

func TestGetServiceNodeStatus(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestServiceNode()
	if err := ctx.InsertServiceNode(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	status, err := ctx.GetServiceNodeStatus(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if status != int(actor.Status) {
		t.Fatal("unequal status")
	}
}

func TestGetServiceNodePauseHeightIfExists(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestServiceNode()
	if err := ctx.InsertServiceNode(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	pauseHeight := 1
	err := ctx.SetServiceNodePauseHeight(actor.Address, int64(pauseHeight))
	if err != nil {
		t.Fatal(err)
	}
	pauseBeforeHeight, err := ctx.GetServiceNodePauseHeightIfExists(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if pauseHeight != int(pauseBeforeHeight) {
		t.Fatalf("incorrect pause height: expected %v, got %v", pauseHeight, pauseBeforeHeight)
	}
}

func TestSetServiceNodesStatusAndUnstakingHeightPausedBefore(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestServiceNode()
	if err := ctx.InsertServiceNode(actor.Address, actor.PublicKey, actor.Output, true, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	pauseBeforeHeight, unstakingHeight, status := int64(1), int64(10), 1
	err := ctx.SetServiceNodesStatusAndUnstakingHeightPausedBefore(pauseBeforeHeight, unstakingHeight, status)
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := ctx.(*PrePersistenceContext).GetServiceNode(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if got.UnstakingHeight != unstakingHeight {
		t.Fatalf("wrong unstaking height: expected %v, got %v", unstakingHeight, got.UnstakingHeight)
	}
	if int(got.Status) != status {
		t.Fatalf("wrong status: expected %v, got %v", status, got.Status)
	}
}

func TestGetServiceNodeOutputAddress(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestServiceNode()
	if err := ctx.InsertServiceNode(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	output, err := ctx.GetServiceNodeOutputAddress(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actor.Output, output) {
		t.Fatalf("incorrect output address expected %v, got %v", actor.Output, output)
	}
}
