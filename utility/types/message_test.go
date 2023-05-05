package types

import (
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/utils"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	defaultTestingChains = []string{"0001"}
	defaultAmountBig     = big.NewInt(1000000)
	defaultAmount        = utils.BigIntToString(defaultAmountBig)
	defaultUnusedLength  = -1
)

func TestMessage_ChangeParameter_ValidateBasic(t *testing.T) {
	owner, err := crypto.GenerateAddress()
	require.NoError(t, err)

	paramKey := "key"
	paramValueRaw := wrapperspb.Int32(1)
	paramValueAny, err := codec.GetCodec().ToAny(paramValueRaw)
	require.NoError(t, err)

	msg := MessageChangeParameter{
		Owner:          owner,
		ParameterKey:   paramKey,
		ParameterValue: paramValueAny,
	}

	err = msg.ValidateBasic()
	require.NoError(t, err)

	msgMissingOwner := proto.Clone(&msg).(*MessageChangeParameter)
	msgMissingOwner.Owner = nil
	require.Equal(t, coreTypes.ErrEmptyAddress().Code(), msgMissingOwner.ValidateBasic().Code())

	msgMissingParamKey := proto.Clone(&msg).(*MessageChangeParameter)
	msgMissingParamKey.ParameterKey = ""
	require.Equal(t, coreTypes.ErrEmptyParamKey().Code(), msgMissingParamKey.ValidateBasic().Code())

	msgMissingParamValue := proto.Clone(&msg).(*MessageChangeParameter)
	msgMissingParamValue.ParameterValue = nil
	require.Equal(t, coreTypes.ErrEmptyParamValue().Code(), msgMissingParamValue.ValidateBasic().Code())
}

func TestMessage_EditStake_ValidateBasic(t *testing.T) {
	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	msg := MessageEditStake{
		ActorType: coreTypes.ActorType_ACTOR_TYPE_APP,
		Address:   addr,
		Chains:    defaultTestingChains,
		Amount:    defaultAmount,
	}
	err = msg.ValidateBasic()
	require.NoError(t, err)

	msgMissingAmount := proto.Clone(&msg).(*MessageEditStake)
	msgMissingAmount.Amount = ""
	er := msgMissingAmount.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyAmount().Code(), er.Code())

	msgInvalidAmount := proto.Clone(&msg).(*MessageEditStake)
	msgInvalidAmount.Amount = "sdk"
	er = msgInvalidAmount.ValidateBasic()
	require.Equal(t, coreTypes.ErrStringToBigInt(er).Code(), er.Code())

	msgEmptyAddress := proto.Clone(&msg).(*MessageEditStake)
	msgEmptyAddress.Address = nil
	er = msgEmptyAddress.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyAddress().Code(), er.Code())

	msgInvalidAddress := proto.Clone(&msg).(*MessageEditStake)
	msgInvalidAddress.Address = []byte("badAddr")
	er = msgInvalidAddress.ValidateBasic()
	expectedErr := coreTypes.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen(defaultUnusedLength))
	require.Equal(t, expectedErr.Code(), er.Code())

	msgEmptyRelayChains := proto.Clone(&msg).(*MessageEditStake)
	msgEmptyRelayChains.Chains = nil
	er = msgEmptyRelayChains.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyRelayChains().Code(), er.Code())

	msgInvalidRelayChains := proto.Clone(&msg).(*MessageEditStake)
	msgInvalidRelayChains.Chains = []string{"notAValidRelayChain"}
	er = msgInvalidRelayChains.ValidateBasic()
	expectedErr = coreTypes.ErrInvalidRelayChainLength(0, relayChainLength)
	require.Equal(t, expectedErr.Code(), er.Code())
}

func TestMessage_Send_ValidateBasic(t *testing.T) {
	addr1, err := crypto.GenerateAddress()
	require.NoError(t, err)

	addr2, err := crypto.GenerateAddress()
	require.NoError(t, err)

	msg := MessageSend{
		FromAddress: addr1,
		ToAddress:   addr2,
		Amount:      defaultAmount,
	}
	er := msg.ValidateBasic()
	require.NoError(t, er)

	msgMissingAddress := proto.Clone(&msg).(*MessageSend)
	msgMissingAddress.FromAddress = nil
	er = msgMissingAddress.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyAddress().Code(), er.Code())

	msgMissingToAddress := proto.Clone(&msg).(*MessageSend)
	msgMissingToAddress.ToAddress = nil
	er = msgMissingToAddress.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyAddress().Code(), er.Code())

	msgMissingAmount := proto.Clone(&msg).(*MessageSend)
	msgMissingAmount.Amount = ""
	er = msgMissingAmount.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyAmount().Code(), er.Code())

	msgInvalidAmount := proto.Clone(&msg).(*MessageSend)
	msgInvalidAmount.Amount = ""
	er = msgInvalidAmount.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyAmount().Code(), er.Code())
}

func TestMessage_Stake_ValidateBasic(t *testing.T) {
	pk, err := crypto.GeneratePublicKey()
	require.NoError(t, err)

	msg := MessageStake{
		ActorType:     coreTypes.ActorType_ACTOR_TYPE_APP,
		PublicKey:     pk.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmount,
		OutputAddress: pk.Address(),
		Signer:        nil,
	}
	er := msg.ValidateBasic()
	require.NoError(t, er)

	msgEmptyPubKey := proto.Clone(&msg).(*MessageStake)
	msgEmptyPubKey.PublicKey = nil
	er = msgEmptyPubKey.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyPublicKey().Code(), er.Code())

	msgEmptyChains := proto.Clone(&msg).(*MessageStake)
	msgEmptyChains.Chains = nil
	er = msgEmptyChains.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyRelayChains().Code(), er.Code())

	msgEmptyAmount := proto.Clone(&msg).(*MessageStake)
	msgEmptyAmount.Amount = ""
	er = msgEmptyAmount.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyAmount().Code(), er.Code())

	msgEmptyOutputAddress := proto.Clone(&msg).(*MessageStake)
	msgEmptyOutputAddress.OutputAddress = nil
	er = msgEmptyOutputAddress.ValidateBasic()
	require.Equal(t, coreTypes.ErrNilOutputAddress().Code(), er.Code())
}

func TestMessage_Unstake_ValidateBasic(t *testing.T) {
	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	msg := MessageUnstake{
		Address: addr,
	}
	er := msg.ValidateBasic()
	require.NoError(t, er)

	msgMissingAddress := proto.Clone(&msg).(*MessageUnstake)
	msgMissingAddress.Address = nil
	er = msgMissingAddress.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyAddress().Code(), er.Code())
}

func TestMessage_Unpause_ValidateBasic(t *testing.T) {
	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	msg := MessageUnpause{
		Address: addr,
	}
	er := msg.ValidateBasic()
	require.NoError(t, er)

	msgMissingAddress := proto.Clone(&msg).(*MessageUnpause)
	msgMissingAddress.Address = nil
	er = msgMissingAddress.ValidateBasic()
	require.Equal(t, coreTypes.ErrEmptyAddress().Code(), er.Code())
}
