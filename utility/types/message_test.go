package types

import (
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	defaultTestingChains = []string{"0001"}
	defaultAmountBig     = big.NewInt(1000000)
	defaultAmount        = BigIntToString(defaultAmountBig)
	defaultUnusedLength  = -1
)

func TestMessage_ChangeParameter_ValidateBasic(t *testing.T) {
	owner, err := crypto.GenerateAddress()
	require.NoError(t, err)

	codec := codec.GetCodec()
	paramKey := "key"
	paramValueRaw := wrapperspb.Int32(1)
	paramValueAny, err := codec.ToAny(paramValueRaw)
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
	require.Equal(t, ErrEmptyAddress().Code(), msgMissingOwner.ValidateBasic().Code())

	msgMissingParamKey := proto.Clone(&msg).(*MessageChangeParameter)
	msgMissingParamKey.ParameterKey = ""
	require.Equal(t, ErrEmptyParamKey().Code(), msgMissingParamKey.ValidateBasic().Code())

	msgMissingParamValue := proto.Clone(&msg).(*MessageChangeParameter)
	msgMissingParamValue.ParameterValue = nil
	require.Equal(t, ErrEmptyParamValue().Code(), msgMissingParamValue.ValidateBasic().Code())
}

func TestMessage_DoubleSign_ValidateBasic(t *testing.T) {
	pk, err := crypto.GeneratePublicKey()
	require.NoError(t, err)

	hashA := crypto.SHA3Hash(pk.Bytes())
	hashB := crypto.SHA3Hash(pk.Address())
	voteA := &LegacyVote{
		PublicKey: pk.Bytes(),
		Height:    1,
		Round:     2,
		Type:      DoubleSignEvidenceType,
		BlockHash: hashA,
	}
	voteB := &LegacyVote{
		PublicKey: pk.Bytes(),
		Height:    1,
		Round:     2,
		Type:      DoubleSignEvidenceType,
		BlockHash: hashB,
	}
	reporter, _ := crypto.GenerateAddress()
	msg := &MessageDoubleSign{
		VoteA:           voteA,
		VoteB:           voteB,
		ReporterAddress: reporter,
	}
	er := msg.ValidateBasic()
	require.NoError(t, er)

	pk2, err := crypto.GeneratePublicKey()
	require.NoError(t, err)
	msgUnequalPubKeys := new(MessageDoubleSign)
	msgUnequalPubKeys.VoteA = proto.Clone(msg.VoteA).(*LegacyVote)
	msgUnequalPubKeys.VoteB = proto.Clone(msg.VoteB).(*LegacyVote)
	msgUnequalPubKeys.VoteA.PublicKey = pk2.Bytes()
	er = msgUnequalPubKeys.ValidateBasic()
	require.Equal(t, ErrUnequalPublicKeys().Code(), er.Code())

	msgUnequalHeights := new(MessageDoubleSign)
	msgUnequalHeights.VoteA = proto.Clone(msg.VoteA).(*LegacyVote)
	msgUnequalHeights.VoteB = proto.Clone(msg.VoteB).(*LegacyVote)
	msgUnequalHeights.VoteA.Height = 2
	er = msgUnequalHeights.ValidateBasic()
	require.Equal(t, ErrUnequalHeights().Code(), er.Code())

	msgUnequalRounds := new(MessageDoubleSign)
	msgUnequalRounds.VoteA = proto.Clone(msg.VoteA).(*LegacyVote)
	msgUnequalRounds.VoteB = proto.Clone(msg.VoteB).(*LegacyVote)
	msgUnequalRounds.VoteA.Round = 1
	er = msgUnequalRounds.ValidateBasic()
	require.Equal(t, ErrUnequalRounds().Code(), er.Code())

	msgEqualVoteHash := new(MessageDoubleSign)
	msgEqualVoteHash.VoteA = proto.Clone(msg.VoteA).(*LegacyVote)
	msgEqualVoteHash.VoteB = proto.Clone(msg.VoteB).(*LegacyVote)
	msgEqualVoteHash.VoteB.BlockHash = hashA
	er = msgEqualVoteHash.ValidateBasic()
	require.Equal(t, ErrEqualVotes().Code(), er.Code())
}

func TestMessage_EditStake_ValidateBasic(t *testing.T) {
	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	msg := MessageEditStake{
		ActorType: ActorType_App,
		Address:   addr,
		Chains:    defaultTestingChains,
		Amount:    defaultAmount,
	}
	err = msg.ValidateBasic()
	require.NoError(t, err)

	msgMissingAmount := proto.Clone(&msg).(*MessageEditStake)
	msgMissingAmount.Amount = ""
	er := msgMissingAmount.ValidateBasic()
	require.Equal(t, ErrEmptyAmount().Code(), er.Code())

	msgInvalidAmount := proto.Clone(&msg).(*MessageEditStake)
	msgInvalidAmount.Amount = "sdk"
	er = msgInvalidAmount.ValidateBasic()
	require.Equal(t, ErrStringToBigInt().Code(), er.Code())

	msgEmptyAddress := proto.Clone(&msg).(*MessageEditStake)
	msgEmptyAddress.Address = nil
	er = msgEmptyAddress.ValidateBasic()
	require.Equal(t, ErrEmptyAddress().Code(), er.Code())

	msgInvalidAddress := proto.Clone(&msg).(*MessageEditStake)
	msgInvalidAddress.Address = []byte("badAddr")
	er = msgInvalidAddress.ValidateBasic()
	expectedErr := ErrInvalidAddressLen(crypto.ErrInvalidAddressLen(defaultUnusedLength))
	require.Equal(t, expectedErr.Code(), er.Code())

	msgEmptyRelayChains := proto.Clone(&msg).(*MessageEditStake)
	msgEmptyRelayChains.Chains = nil
	er = msgEmptyRelayChains.ValidateBasic()
	require.Equal(t, ErrEmptyRelayChains().Code(), er.Code())

	msgInvalidRelayChains := proto.Clone(&msg).(*MessageEditStake)
	msgInvalidRelayChains.Chains = []string{"notAValidRelayChain"}
	er = msgInvalidRelayChains.ValidateBasic()
	expectedErr = ErrInvalidRelayChainLength(0, RelayChainLength)
	require.Equal(t, expectedErr.Code(), er.Code())
}

func TestMessageSend_ValidateBasic(t *testing.T) {
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
	require.Equal(t, ErrEmptyAddress().Code(), er.Code())

	msgMissingToAddress := proto.Clone(&msg).(*MessageSend)
	msgMissingToAddress.ToAddress = nil
	er = msgMissingToAddress.ValidateBasic()
	require.Equal(t, ErrEmptyAddress().Code(), er.Code())

	msgMissingAmount := proto.Clone(&msg).(*MessageSend)
	msgMissingAmount.Amount = ""
	er = msgMissingAmount.ValidateBasic()
	require.Equal(t, ErrEmptyAmount().Code(), er.Code())

	msgInvalidAmount := proto.Clone(&msg).(*MessageSend)
	msgInvalidAmount.Amount = ""
	er = msgInvalidAmount.ValidateBasic()
	require.Equal(t, ErrEmptyAmount().Code(), er.Code())
}

func TestMessageStake_ValidateBasic(t *testing.T) {
	pk, err := crypto.GeneratePublicKey()
	require.NoError(t, err)

	msg := MessageStake{
		ActorType:     ActorType_App,
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
	require.Equal(t, ErrEmptyPublicKey().Code(), er.Code())

	msgEmptyChains := proto.Clone(&msg).(*MessageStake)
	msgEmptyChains.Chains = nil
	er = msgEmptyChains.ValidateBasic()
	require.Equal(t, ErrEmptyRelayChains().Code(), er.Code())

	msgEmptyAmount := proto.Clone(&msg).(*MessageStake)
	msgEmptyAmount.Amount = ""
	er = msgEmptyAmount.ValidateBasic()
	require.Equal(t, ErrEmptyAmount().Code(), er.Code())

	msgEmptyOutputAddress := proto.Clone(&msg).(*MessageStake)
	msgEmptyOutputAddress.OutputAddress = nil
	er = msgEmptyOutputAddress.ValidateBasic()
	require.Equal(t, ErrNilOutputAddress().Code(), er.Code())
}

func TestMessageUnstake_ValidateBasic(t *testing.T) {
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
	require.Equal(t, ErrEmptyAddress().Code(), er.Code())
}

func TestMessageUnpause_ValidateBasic(t *testing.T) {
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
	require.Equal(t, ErrEmptyAddress().Code(), er.Code())
}

func TestRelayChain_Validate(t *testing.T) {
	relayChainValid := RelayChain("0001")
	err := relayChainValid.Validate()
	require.NoError(t, err)

	relayChainInvalidLength := RelayChain("001")
	expectedError := ErrInvalidRelayChainLength(0, RelayChainLength)
	err = relayChainInvalidLength.Validate()
	require.Equal(t, expectedError.Code(), err.Code())

	relayChainEmpty := RelayChain("")
	expectedError = ErrEmptyRelayChain()
	err = relayChainEmpty.Validate()
	require.Equal(t, expectedError.Code(), err.Code())
}
