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

func NewTestValidator() typesGenesis.Validator {
	pub1, _ := crypto.GeneratePublicKey()
	addr1 := pub1.Address()
	addr2, _ := crypto.GenerateAddress()
	return typesGenesis.Validator{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		ServiceUrl:      typesGenesis.DefaultServiceUrl,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    0,
		UnstakingHeight: 0,
		Output:          addr2,
	}
}

func TestGetValidatorExists(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestValidator()
	addr2, _ := crypto.GenerateAddress()
	if err := ctx.InsertValidator(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	exists, err := ctx.GetValidatorExists(actor.Address)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exists does not")
	}
	exists, err = ctx.GetValidatorExists(addr2)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that exists should not")
	}
}

func TestGetValidator(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestValidator()
	if err := ctx.InsertValidator(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	got, _, err := ctx.(*PrePersistenceContext).GetValidator(actor.Address)
	require.NoError(t, err)
	if !bytes.Equal(actor.Address, got.Address) || !bytes.Equal(actor.PublicKey, got.PublicKey) {
		t.Fatalf("unexpected actor returned; expected %v got %v", actor, got)
	}
}

func TestGetAllValidators(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor1 := NewTestValidator()
	actor2 := NewTestValidator()
	if err := ctx.InsertValidator(actor1.Address, actor1.PublicKey, actor1.Output, actor1.Paused, int(actor1.Status),
		actor1.ServiceUrl, actor1.StakedTokens, int64(actor1.PausedHeight), actor1.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := ctx.InsertValidator(actor2.Address, actor2.PublicKey, actor2.Output, actor2.Paused, int(actor2.Status),
		actor2.ServiceUrl, actor2.StakedTokens, int64(actor2.PausedHeight), actor2.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	validators, err := ctx.(*PrePersistenceContext).GetAllValidators(0)
	require.NoError(t, err)
	got1, got2 := false, false
	for _, a := range validators {
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

func TestUpdateValidator(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestValidator()
	if err := ctx.InsertValidator(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	bigExpectedTokens := big.NewInt(1)
	one := types.BigIntToString(bigExpectedTokens)
	before, _, err := ctx.(*PrePersistenceContext).GetValidator(actor.Address)
	require.NoError(t, err)
	tokens := before.StakedTokens
	bigBeforeTokens, err := types.StringToBigInt(tokens)
	require.NoError(t, err)
	err = ctx.UpdateValidator(actor.Address, typesGenesis.DefaultServiceUrl, one)
	require.NoError(t, err)
	got, _, err := ctx.(*PrePersistenceContext).GetValidator(actor.Address)
	require.NoError(t, err)
	bigAfterTokens, err := types.StringToBigInt(got.StakedTokens)
	require.NoError(t, err)
	bigAfterTokens.Sub(bigAfterTokens, bigBeforeTokens)
	if bigAfterTokens.Cmp(bigExpectedTokens) != 0 {
		t.Fatal("incorrect after balance")
	}
}

func TestDeleteValidator(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestValidator()
	if err := ctx.InsertValidator(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	err := ctx.DeleteValidator(actor.Address)
	require.NoError(t, err)
	exists, err := ctx.(*PrePersistenceContext).GetValidatorExists(actor.Address)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor exists when it shouldn't")
	}
}

func TestGetValidatorsReadyToUnstake(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestValidator()
	if err := ctx.InsertValidator(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := ctx.SetValidatorUnstakingHeightAndStatus(actor.Address, 0, 1); err != nil {
		t.Fatal(err)
	}
	unstakingValidators, err := ctx.(*PrePersistenceContext).GetValidatorsReadyToUnstake(0, 1)
	require.NoError(t, err)
	if !bytes.Equal(unstakingValidators[0].Address, actor.Address) {
		t.Fatalf("wrong actor returned, expected addr %v, got %v", unstakingValidators[0].Address, actor.Address)
	}
}

func TestGetValidatorStatus(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestValidator()
	if err := ctx.InsertValidator(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	status, err := ctx.GetValidatorStatus(actor.Address)
	require.NoError(t, err)
	if status != int(actor.Status) {
		t.Fatal("unequal status")
	}
}

func TestGetValidatorPauseHeightIfExists(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestValidator()
	if err := ctx.InsertValidator(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	pausedHeight := 1
	err := ctx.SetValidatorPauseHeight(actor.Address, int64(pausedHeight))
	require.NoError(t, err)
	pauseBeforeHeight, err := ctx.GetValidatorPauseHeightIfExists(actor.Address)
	require.NoError(t, err)
	if pausedHeight != int(pauseBeforeHeight) {
		t.Fatalf("incorrect pause height: expected %v, got %v", pausedHeight, pauseBeforeHeight)
	}
}

func TestSetValidatorsStatusAndUnstakingHeightIfPausedBefore(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestValidator()
	if err := ctx.InsertValidator(actor.Address, actor.PublicKey, actor.Output, true, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	pauseBeforeHeight, unstakingHeight, status := int64(1), int64(10), 1
	err := ctx.SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pauseBeforeHeight, unstakingHeight, status)
	require.NoError(t, err)
	got, _, err := ctx.(*PrePersistenceContext).GetValidator(actor.Address)
	require.NoError(t, err)
	if got.UnstakingHeight != unstakingHeight {
		t.Fatalf("wrong unstaking height: expected %v, got %v", unstakingHeight, got.UnstakingHeight)
	}
	if int(got.Status) != status {
		t.Fatalf("wrong status: expected %v, got %v", status, got.Status)
	}
}

func TestGetValidatorOutputAddress(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestValidator()
	if err := ctx.InsertValidator(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.ServiceUrl, actor.StakedTokens, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	output, err := ctx.GetValidatorOutputAddress(actor.Address)
	require.NoError(t, err)
	if !bytes.Equal(actor.Output, output) {
		t.Fatalf("incorrect output address expected %v, got %v", actor.Output, output)
	}
}
