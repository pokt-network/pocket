package types

import (
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"math/big"
	"testing"
)

var (
	defaultTestingChains = []string{"0001"}
	defaultServiceURL    = "https://foo.bar:443"
	defaultAmountBig     = big.NewInt(1000000)
	defaultAmount        = types.BigIntToString(defaultAmountBig)
	defaultFeeBig        = big.NewInt(10000)
	defaultFee           = types.BigIntToString(defaultFeeBig)
)

func TestMessageChangeParameter_ValidateBasic(t *testing.T) {
	codec := UtilityCodec()
	owner, _ := crypto.GenerateAddress()
	paramKey := "key"
	paramValueRaw := wrapperspb.Int32(1)
	paramValueAny, err := codec.ToAny(paramValueRaw)
	if err != nil {
		t.Fatal(err)
	}
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

func TestMessageEditStakeApp_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageEditStakeApp{
		Address:     addr,
		Chains:      defaultTestingChains,
		AmountToAdd: defaultAmount,
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgMissingAmount := msg
	msgMissingAmount.AmountToAdd = ""
	if err := msgMissingAmount.ValidateBasic(); err.Code() != types.ErrEmptyAmount().Code() {
		t.Fatal(err)
	}
	msgInvalidAmount := msg
	msgInvalidAmount.AmountToAdd = "sdk"
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
	if err := msgInvalidAddress.ValidateBasic(); err.Code() != types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen()).Code() {
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

func TestMessageEditStakeFisherman_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageEditStakeFisherman{
		Address:     addr,
		Chains:      defaultTestingChains,
		AmountToAdd: defaultAmount,
		ServiceURL:  defaultServiceURL,
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgMissingAmount := msg
	msgMissingAmount.AmountToAdd = ""
	if err := msgMissingAmount.ValidateBasic(); err.Code() != types.ErrEmptyAmount().Code() {
		t.Fatal(err)
	}
	msgInvalidAmount := msg
	msgInvalidAmount.AmountToAdd = "sdk"
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
	if err := msgInvalidAddress.ValidateBasic(); err.Code() != types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen()).Code() {
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
	msgEmptyServiceURL := msg
	msgEmptyServiceURL.ServiceURL = ""
	if err := msgEmptyServiceURL.ValidateBasic(); err.Code() != types.ErrInvalidServiceURL("").Code() {
		t.Fatal(err)
	}
}

func TestMessageEditStakeServiceNode_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageEditStakeServiceNode{
		Address:     addr,
		Chains:      defaultTestingChains,
		AmountToAdd: defaultAmount,
		ServiceURL:  defaultServiceURL,
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgMissingAmount := msg
	msgMissingAmount.AmountToAdd = ""
	if err := msgMissingAmount.ValidateBasic(); err.Code() != types.ErrEmptyAmount().Code() {
		t.Fatal(err)
	}
	msgInvalidAmount := msg
	msgInvalidAmount.AmountToAdd = "sdk"
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
	if err := msgInvalidAddress.ValidateBasic(); err.Code() != types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen()).Code() {
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
	msgEmptyServiceURL := msg
	msgEmptyServiceURL.ServiceURL = ""
	if err := msgEmptyServiceURL.ValidateBasic(); err.Code() != types.ErrInvalidServiceURL("").Code() {
		t.Fatal(err)
	}
}

func TestMessageEditStakeValidator_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageEditStakeValidator{
		Address:     addr,
		AmountToAdd: defaultAmount,
		ServiceURL:  defaultServiceURL,
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgMissingAmount := msg
	msgMissingAmount.AmountToAdd = ""
	if err := msgMissingAmount.ValidateBasic(); err.Code() != types.ErrEmptyAmount().Code() {
		t.Fatal(err)
	}
	msgInvalidAmount := msg
	msgInvalidAmount.AmountToAdd = "sdk"
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
	if err := msgInvalidAddress.ValidateBasic(); err.Code() != types.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen()).Code() {
		t.Fatal(err)
	}
	msgEmptyServiceURL := msg
	msgEmptyServiceURL.ServiceURL = ""
	if err := msgEmptyServiceURL.ValidateBasic(); err.Code() != types.ErrInvalidServiceURL("").Code() {
		t.Fatal(err)
	}
}

func TestMessageFishermanPauseServiceNode_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageFishermanPauseServiceNode{
		Address:  addr,
		Reporter: addr,
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	msgMissingReporter := msg
	msgMissingReporter.Reporter = nil
	if err := msgMissingReporter.ValidateBasic(); err.Code() != types.ErrEmptyAddress().Code() {
		t.Fatal(err)
	}
	msgMissingAddress := msg
	msgMissingAddress.Address = nil
	if err := msgMissingAddress.ValidateBasic(); err.Code() != types.ErrEmptyAddress().Code() {
		t.Fatal(err)
	}
}

func TestMessagePauseApp_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessagePauseApp{
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

func TestMessagePauseFisherman_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessagePauseFisherman{
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

func TestMessagePauseServiceNode_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessagePauseServiceNode{
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

func TestMessagePauseValidator_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessagePauseValidator{
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

func TestMessageStakeApp_ValidateBasic(t *testing.T) {
	pk, _ := crypto.GeneratePublicKey()
	msg := MessageStakeApp{
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

func TestMessageStakeFisherman_ValidateBasic(t *testing.T) {
	pk, _ := crypto.GeneratePublicKey()
	msg := MessageStakeFisherman{
		PublicKey:     pk.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmount,
		ServiceURL:    defaultServiceURL,
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
	msgEmptyServiceUrl := msg
	msgEmptyServiceUrl.ServiceURL = ""
	if err := msgEmptyServiceUrl.ValidateBasic(); err.Code() != types.ErrInvalidServiceURL("").Code() {
		t.Fatal(err)
	}
}

func TestMessageStakeServiceNode_ValidateBasic(t *testing.T) {
	pk, _ := crypto.GeneratePublicKey()
	msg := MessageStakeServiceNode{
		PublicKey:     pk.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmount,
		ServiceURL:    defaultServiceURL,
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
	msgEmptyServiceUrl := msg
	msgEmptyServiceUrl.ServiceURL = ""
	if err := msgEmptyServiceUrl.ValidateBasic(); err.Code() != types.ErrInvalidServiceURL("").Code() {
		t.Fatal(err)
	}
}

func TestMessageStakeValidator_ValidateBasic(t *testing.T) {
	pk, _ := crypto.GeneratePublicKey()
	msg := MessageStakeValidator{
		PublicKey:     pk.Bytes(),
		Amount:        defaultAmount,
		ServiceURL:    defaultServiceURL,
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
	msgEmptyServiceUrl := msg
	msgEmptyServiceUrl.ServiceURL = ""
	if err := msgEmptyServiceUrl.ValidateBasic(); err.Code() != types.ErrInvalidServiceURL("").Code() {
		t.Fatal(err)
	}
}

func TestMessageUnpauseApp_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageUnpauseApp{
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

func TestMessageUnpauseFisherman_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageUnpauseFisherman{
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

func TestMessageUnpauseServiceNode_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageUnpauseServiceNode{
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

func TestMessageUnpauseValidator_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageUnpauseValidator{
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

func TestMessageUnstakeApp_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageUnstakeApp{
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

func TestMessageUnstakeFisherman_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageUnstakeFisherman{
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

func TestMessageUnstakeServiceNode_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageUnstakeServiceNode{
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

func TestMessageUnstakeValidator_ValidateBasic(t *testing.T) {
	addr, _ := crypto.GenerateAddress()
	msg := MessageUnstakeValidator{
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
