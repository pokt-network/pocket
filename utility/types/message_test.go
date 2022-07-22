package types

import (
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	defaultTestingChains = []string{"0001"}
	defaultServiceUrl    = "https://foo.bar:443"
	defaultAmountBig     = big.NewInt(1000000)
	defaultAmount        = types.BigIntToString(defaultAmountBig)
	defaultFeeBig        = big.NewInt(10000)
	defaultFee           = types.BigIntToString(defaultFeeBig)
	defaultUnusedLength  = -1
)

func TestMessageChangeParameter_ValidateBasic(t *testing.T) {
	codec := types.GetCodec()
	owner, _ := crypto.GenerateAddress()
	paramKey := "key"
	paramValueRaw := wrapperspb.Int32(1)
	paramValueAny, err := codec.ToAny(paramValueRaw)
	require.NoError(t, err)
	msg := MessageChangeParameter{
		Owner:          owner,
		ParameterKey:   paramKey,
		ParameterValue: paramValueAny,
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgMissingOwner := msg
	msgMissingOwner.Owner = nil
	if err := msgMissingOwner.ValidateBasic(); err.Code() != types.ErrEmptyAddress().Code() {
		t.Fatal(err)
	}
	msgMissingParamKey := msg
	msgMissingParamKey.ParameterKey = ""
	if err := msgMissingParamKey.ValidateBasic(); err.Code() != types.ErrEmptyParamKey().Code() {
		t.Fatal(err)
	}
	msgMissingParamValue := msg
	msgMissingParamValue.ParameterValue = nil
	if err := msgMissingParamValue.ValidateBasic(); err.Code() != types.ErrEmptyParamValue().Code() {
		t.Fatal(err)
	}
}

func TestMessageDoubleSign_ValidateBasic(t *testing.T) {
	pk, _ := crypto.GeneratePublicKey()
	hashA := crypto.SHA3Hash(pk.Bytes())
	hashB := crypto.SHA3Hash(pk.Address())
	voteA := &Vote{
		PublicKey: pk.Bytes(),
		Height:    1,
		Round:     2,
		Type:      DoubleSignEvidenceType,
		BlockHash: hashA,
	}
	voteB := &Vote{
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
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgUnequalPubKeys := new(MessageDoubleSign)
	msgUnequalPubKeys.VoteA = new(Vote)
	msgUnequalPubKeys.VoteB = new(Vote)
	*msgUnequalPubKeys.VoteA = *msg.VoteA
	*msgUnequalPubKeys.VoteB = *msg.VoteB
	pk2, _ := crypto.GeneratePublicKey()
	msgUnequalPubKeys.VoteA.PublicKey = pk2.Bytes()
	if err := msgUnequalPubKeys.ValidateBasic(); err.Code() != types.ErrUnequalPublicKeys().Code() {
		t.Fatal(err)
	}
	msgUnequalHeights := new(MessageDoubleSign)
	msgUnequalHeights.VoteA = new(Vote)
	msgUnequalHeights.VoteB = new(Vote)
	*msgUnequalHeights.VoteA = *msg.VoteA
	*msgUnequalHeights.VoteB = *msg.VoteB
	msgUnequalHeights.VoteA.Height = 2
	if err := msgUnequalHeights.ValidateBasic(); err.Code() != types.ErrUnequalHeights().Code() {
		t.Fatal(err)
	}
	msgUnequalRounds := new(MessageDoubleSign)
	msgUnequalRounds.VoteA = new(Vote)
	msgUnequalRounds.VoteB = new(Vote)
	*msgUnequalRounds.VoteA = *msg.VoteA
	*msgUnequalRounds.VoteB = *msg.VoteB
	msgUnequalRounds.VoteA.Round = 1
	if err := msgUnequalRounds.ValidateBasic(); err.Code() != types.ErrUnequalRounds().Code() {
		t.Fatal(err)
	}
	//msgUnequalVoteTypes := new(MessageDoubleSign) TODO only one type of evidence right now
	//msgUnequalVoteTypes.VoteA = new(Vote)
	//msgUnequalVoteTypes.VoteB = new(Vote)
	//*msgUnequalVoteTypes.VoteA = *msg.VoteA
	//*msgUnequalVoteTypes.VoteB = *msg.VoteB
	//msgUnequalVoteTypes.VoteA.Type = 0
	//if err := msgUnequalVoteTypes.ValidateBasic(); err.Code() != types.ErrUnequalVoteTypes().Code() {
	//	t.Fatal(err)
	//}
	msgEqualVoteHash := new(MessageDoubleSign)
	msgEqualVoteHash.VoteA = new(Vote)
	msgEqualVoteHash.VoteB = new(Vote)
	*msgEqualVoteHash.VoteA = *msg.VoteA
	*msgEqualVoteHash.VoteB = *msg.VoteB
	msgEqualVoteHash.VoteB.BlockHash = hashA
	if err := msgEqualVoteHash.ValidateBasic(); err.Code() != types.ErrEqualVotes().Code() {
		t.Fatal(err)
	}
}

func TestMessageEditStake_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageEditStake{
		Address: addr,
		Chains:  defaultTestingChains,
		Amount:  defaultAmount,
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgMissingAmount := msg
	msgMissingAmount.Amount = ""
	if err := msgMissingAmount.ValidateBasic(); err.Code() != types.ErrEmptyAmount().Code() {
		t.Fatal(err)
	}
	msgInvalidAmount := msg
	msgInvalidAmount.Amount = "sdk"
	if err := msgInvalidAmount.ValidateBasic(); err.Code() != types.ErrStringToBigInt().Code() {
		t.Fatal(err)
	}
	msgEmptyAddress := msg
	msgEmptyAddress.Address = nil
	if err := msgEmptyAddress.ValidateBasic(); err.Code() != types.ErrEmptyAddress().Code() {
		t.Fatal(err)
	}
	msgInvalidAddress := msg
	msgInvalidAddress.Address = []byte("badAddr")
	if err := msgInvalidAddress.ValidateBasic(); err.Code() != types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen(defaultUnusedLength)).Code() {
		t.Fatal(err)
	}
	msgEmptyRelayChains := msg
	msgEmptyRelayChains.Chains = nil
	if err := msgEmptyRelayChains.ValidateBasic(); err.Code() != types.ErrEmptyRelayChains().Code() {
		t.Fatal(err)
	}
	msgInvalidRelayChains := msg
	msgInvalidRelayChains.Chains = []string{"notAValidRelayChain"}
	if err := msgInvalidRelayChains.ValidateBasic(); err.Code() != types.ErrInvalidRelayChainLength(0, RelayChainLength).Code() {
		t.Fatal(err)
	}
}

func TestMessageSend_ValidateBasic(t *testing.T) {
	addr1, _ := crypto.GenerateAddress()
	addr2, _ := crypto.GenerateAddress()
	msg := MessageSend{
		FromAddress: addr1,
		ToAddress:   addr2,
		Amount:      defaultAmount,
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgMissingAddress := msg
	msgMissingAddress.FromAddress = nil
	if err := msgMissingAddress.ValidateBasic(); err.Code() != types.ErrEmptyAddress().Code() {
		t.Fatal(err)
	}
	msgMissingToAddress := msg
	msgMissingToAddress.ToAddress = nil
	if err := msgMissingToAddress.ValidateBasic(); err.Code() != types.ErrEmptyAddress().Code() {
		t.Fatal(err)
	}
	msgMissingAmount := msg
	msgMissingAmount.Amount = ""
	if err := msgMissingAmount.ValidateBasic(); err.Code() != types.ErrEmptyAmount().Code() {
		t.Fatal(err)
	}
	msgInvalidAmount := msg
	msgInvalidAmount.Amount = ""
	if err := msgInvalidAmount.ValidateBasic(); err.Code() != types.ErrEmptyAmount().Code() {
		t.Fatal(err)
	}
}

func TestMessageStake_ValidateBasic(t *testing.T) {
	pk, _ := crypto.GeneratePublicKey()
	msg := MessageStake{
		PublicKey:     pk.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmount,
		OutputAddress: pk.Address(),
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgEmptyPubKey := msg
	msgEmptyPubKey.PublicKey = nil
	if err := msgEmptyPubKey.ValidateBasic(); err.Code() != types.ErrEmptyPublicKey().Code() {
		t.Fatal(err)
	}
	msgEmptyChains := msg
	msgEmptyChains.Chains = nil
	if err := msgEmptyChains.ValidateBasic(); err.Code() != types.ErrEmptyRelayChains().Code() {
		t.Fatal(err)
	}
	msgEmptyAmount := msg
	msgEmptyAmount.Amount = ""
	if err := msgEmptyAmount.ValidateBasic(); err.Code() != types.ErrEmptyAmount().Code() {
		t.Fatal(err)
	}
	msgEmptyOutputAddress := msg
	msgEmptyOutputAddress.OutputAddress = nil
	if err := msgEmptyOutputAddress.ValidateBasic(); err.Code() != types.ErrNilOutputAddress().Code() {
		t.Fatal(err)
	}
}

func TestMessageUnpause_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageUnpause{
		Address: addr,
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgMissingAddress := msg
	msgMissingAddress.Address = nil
	if err := msgMissingAddress.ValidateBasic(); err.Code() != types.ErrEmptyAddress().Code() {
		t.Fatal(err)
	}
}

func TestMessageUnstake_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageUnstake{
		Address: addr,
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgMissingAddress := msg
	msgMissingAddress.Address = nil
	if err := msgMissingAddress.ValidateBasic(); err.Code() != types.ErrEmptyAddress().Code() {
		t.Fatal(err)
	}
}

func TestRelayChain_Validate(t *testing.T) {
	relayChainValid := RelayChain("0001")
	relayChainInvalidLength := RelayChain("001")
	relayChainEmpty := RelayChain("")
	if err := relayChainValid.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := relayChainInvalidLength.Validate(); err.Code() != types.ErrInvalidRelayChainLength(0, RelayChainLength).Code() {
		t.Fatal(err)
	}
	if err := relayChainEmpty.Validate(); err.Code() != types.ErrEmptyRelayChain().Code() {
		t.Fatal(err)
	}
}
