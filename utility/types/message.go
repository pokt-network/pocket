package types

import (
	"bytes"
	"net/url"
	"strconv"
	"strings"

	crypto2 "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/proto"
)

type Message interface {
	proto.Message
	SetSigner(signer []byte)
	ValidateBasic() types.Error
}

var (
	_ Message = &MessageStakeApp{}
)

func (x *MessageStakeApp) ValidateBasic() types.Error {
	// validate amount
	if err := ValidateAmount(x.Amount); err != nil {
		return err
	}
	if err := ValidatePublicKey(x.PublicKey); err != nil {
		return err
	}
	if err := ValidateRelayChains(x.Chains); err != nil {
		return err
	}
	// validate output address
	return ValidateOutputAddress(x.OutputAddress)
}

func (x *MessageStakeApp) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageEditStakeApp) ValidateBasic() types.Error {
	// validate amount
	if err := ValidateAmount(x.AmountToAdd); err != nil {
		return err
	}
	if err := ValidateAddress(x.Address); err != nil {
		return err
	}
	if err := ValidateRelayChains(x.Chains); err != nil {
		return err
	}
	return nil
}

func (x *MessageEditStakeApp) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnstakeApp) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessageUnstakeApp) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnpauseApp) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessageUnpauseApp) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessagePauseApp) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessagePauseApp) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageStakeServiceNode) ValidateBasic() types.Error {
	// validate amount
	if err := ValidateAmount(x.Amount); err != nil {
		return err
	}
	if err := ValidatePublicKey(x.PublicKey); err != nil {
		return err
	}
	if err := ValidateRelayChains(x.Chains); err != nil {
		return err
	}
	if err := ValidateServiceUrl(x.ServiceUrl); err != nil {
		return err
	}
	// validate output address
	return ValidateOutputAddress(x.OutputAddress)
}

func (x *MessageStakeServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageEditStakeServiceNode) ValidateBasic() types.Error {
	// validate amount
	if err := ValidateAmount(x.AmountToAdd); err != nil {
		return err
	}
	if err := ValidateAddress(x.Address); err != nil {
		return err
	}
	if err := ValidateRelayChains(x.Chains); err != nil {
		return err
	}
	if err := ValidateServiceUrl(x.ServiceUrl); err != nil {
		return err
	}
	return nil
}

func (x *MessageEditStakeServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnstakeServiceNode) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessageUnstakeServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnpauseServiceNode) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessageUnpauseServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessagePauseServiceNode) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessagePauseServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageStakeFisherman) ValidateBasic() types.Error {
	// validate amount
	if err := ValidateAmount(x.Amount); err != nil {
		return err
	}
	if err := ValidatePublicKey(x.PublicKey); err != nil {
		return err
	}
	if err := ValidateRelayChains(x.Chains); err != nil {
		return err
	}
	if err := ValidateServiceUrl(x.ServiceUrl); err != nil {
		return err
	}
	// validate output address
	return ValidateOutputAddress(x.OutputAddress)
}

func (x *MessageChangeParameter) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageChangeParameter) ValidateBasic() types.Error {
	if x.ParameterKey == "" {
		return types.ErrEmptyParamKey()
	}
	if x.ParameterValue == nil {
		return types.ErrEmptyParamValue()
	}
	if err := ValidateAddress(x.Owner); err != nil {
		return err
	}
	return nil
}

func (x *MessageStakeFisherman) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageEditStakeFisherman) ValidateBasic() types.Error {
	// validate amount
	if err := ValidateAmount(x.AmountToAdd); err != nil {
		return err
	}
	if err := ValidateAddress(x.Address); err != nil {
		return err
	}
	if err := ValidateRelayChains(x.Chains); err != nil {
		return err
	}
	if err := ValidateServiceUrl(x.ServiceUrl); err != nil {
		return err
	}
	return nil
}

func (x *MessageEditStakeFisherman) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnstakeFisherman) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessageUnstakeFisherman) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnpauseFisherman) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessageUnpauseFisherman) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessagePauseFisherman) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessagePauseFisherman) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageFishermanPauseServiceNode) ValidateBasic() types.Error {
	if err := ValidateAddress(x.Reporter); err != nil {
		return err
	}
	return ValidateAddress(x.Address)
}

func (x *MessageFishermanPauseServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageStakeValidator) ValidateBasic() types.Error {
	// validate amount
	if err := ValidateAmount(x.Amount); err != nil {
		return err
	}
	if err := ValidatePublicKey(x.PublicKey); err != nil {
		return err
	}
	if err := ValidateServiceUrl(x.ServiceUrl); err != nil {
		return err
	}
	// validate output address
	return ValidateOutputAddress(x.OutputAddress)
}

func (x *MessageStakeValidator) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageEditStakeValidator) ValidateBasic() types.Error {
	// validate amount
	if err := ValidateAmount(x.AmountToAdd); err != nil {
		return err
	}
	if err := ValidateAddress(x.Address); err != nil {
		return err
	}
	if err := ValidateServiceUrl(x.ServiceUrl); err != nil {
		return err
	}
	return nil
}

func (x *MessageEditStakeValidator) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnstakeValidator) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessageUnstakeValidator) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnpauseValidator) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessageUnpauseValidator) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessagePauseValidator) ValidateBasic() types.Error {
	return ValidateAddress(x.Address)
}

func (x *MessagePauseValidator) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageDoubleSign) ValidateBasic() types.Error {
	if err := x.VoteA.ValidateBasic(); err != nil {
		return err
	}
	if err := x.VoteB.ValidateBasic(); err != nil {
		return err
	}
	if !bytes.Equal(x.VoteA.PublicKey, x.VoteB.PublicKey) {
		return types.ErrUnequalPublicKeys()
	}
	if x.VoteA.Type != x.VoteB.Type {
		return types.ErrUnequalVoteTypes()
	}
	if x.VoteA.Height != x.VoteB.Height {
		return types.ErrUnequalHeights()
	}
	if x.VoteA.Round != x.VoteB.Round {
		return types.ErrUnequalRounds()
	}
	if bytes.Equal(x.VoteA.BlockHash, x.VoteB.BlockHash) {
		return types.ErrEqualVotes()
	}
	return nil
}

func (x *MessageDoubleSign) SetSigner(signer []byte) {
	x.ReporterAddress = signer
}

func (x *MessageSend) ValidateBasic() types.Error {
	if err := ValidateAddress(x.FromAddress); err != nil {
		return err
	}
	if err := ValidateAddress(x.ToAddress); err != nil {
		return err
	}
	if err := ValidateAmount(x.Amount); err != nil {
		return err
	}
	return nil
}

func (x *MessageSend) SetSigner(signer []byte) {
	// no op
}

func ValidateAddress(address []byte) types.Error {
	if address == nil {
		return types.ErrEmptyAddress()
	}
	if len(address) != crypto2.AddressLen {
		return types.ErrInvalidAddressLen(crypto2.ErrInvalidAddressLen())
	}
	return nil
}

func ValidateOutputAddress(address []byte) types.Error {
	if address == nil {
		return types.ErrNilOutputAddress()
	}
	if len(address) != crypto2.AddressLen {
		return types.ErrInvalidAddressLen(crypto2.ErrInvalidAddressLen())
	}
	return nil
}

func ValidatePublicKey(publicKey []byte) types.Error {
	// validate public key
	if publicKey == nil {
		return types.ErrEmptyPublicKey()
	}
	if len(publicKey) != crypto2.PublicKeyLen {
		return types.ErrInvalidPublicKeylen(crypto2.ErrInvalidPublicKeyLen())
	}
	return nil
}

func ValidateHash(hash []byte) types.Error {
	if hash == nil {
		return types.ErrEmptyHash()
	}
	if len(hash) != crypto2.SHA3HashLen {
		return types.ErrInvalidHashLength(crypto2.ErrInvalidHashLen())
	}
	return nil
}

func ValidateRelayChains(chains []string) types.Error {
	if chains == nil {
		return types.ErrEmptyRelayChains()
	}
	for i := 0; i < len(chains); i++ {
		relayChain := RelayChain(chains[i])
		if err := relayChain.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func ValidateAmount(amount string) types.Error {
	if amount == "" {
		return types.ErrEmptyAmount()
	}
	if _, err := types.StringToBigInt(amount); err != nil {
		return err
	}
	return nil
}

func ValidateServiceUrl(u string) types.Error {
	u = strings.ToLower(u)
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return types.ErrInvalidServiceUrl(err.Error())
	}
	if u[:8] != HttpsPrefix && u[:7] != HttpPrefix {
		return types.ErrInvalidServiceUrl(InvalidURLPrefix)
	}
	temp := strings.Split(u, Colon)
	if len(temp) != 3 {
		return types.ErrInvalidServiceUrl(PortRequired)
	}
	port, err := strconv.Atoi(temp[2])
	if err != nil {
		return types.ErrInvalidServiceUrl(NonNumberPort)
	}
	if port > 65535 || port < 0 {
		return types.ErrInvalidServiceUrl(PortOutOfRange)
	}
	if !strings.Contains(u, Period) {
		return types.ErrInvalidServiceUrl(NoPeriod)
	}
	return nil
}
