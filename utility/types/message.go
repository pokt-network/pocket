package types

import (
	"bytes"
	"log"
	"net/url"
	"strconv"
	"strings"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/proto"
)

type Message interface {
	proto.Message

	SetSigner(signer []byte)
	ValidateBasic() types.Error
}

var _ Message = &MessageStakeApp{}

func (msg *MessageStakeApp) ValidateBasic() types.Error {
	if err := ValidateAmount(msg.Amount); err != nil {
		return err
	}
	if err := ValidatePublicKey(msg.PublicKey); err != nil {
		return err
	}
	if err := ValidateRelayChains(msg.Chains); err != nil {
		return err
	}
	return ValidateOutputAddress(msg.OutputAddress)
}

func (msg *MessageStakeApp) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageEditStakeApp) ValidateBasic() types.Error {
	if err := ValidateAmount(msg.AmountToAdd); err != nil {
		return err
	}
	if err := ValidateAddress(msg.Address); err != nil {
		return err
	}
	if err := ValidateRelayChains(msg.Chains); err != nil {
		return err
	}
	return nil
}

func (msg *MessageEditStakeApp) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageUnstakeApp) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessageUnstakeApp) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageUnpauseApp) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessageUnpauseApp) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessagePauseApp) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessagePauseApp) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageStakeServiceNode) ValidateBasic() types.Error {
	if err := ValidateAmount(msg.Amount); err != nil {
		return err
	}
	if err := ValidatePublicKey(msg.PublicKey); err != nil {
		return err
	}
	if err := ValidateRelayChains(msg.Chains); err != nil {
		return err
	}
	if err := ValidateServiceUrl(msg.ServiceUrl); err != nil {
		return err
	}
	return ValidateOutputAddress(msg.OutputAddress)
}

func (msg *MessageStakeServiceNode) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageEditStakeServiceNode) ValidateBasic() types.Error {
	if err := ValidateAmount(msg.AmountToAdd); err != nil {
		return err
	}
	if err := ValidateAddress(msg.Address); err != nil {
		return err
	}
	if err := ValidateRelayChains(msg.Chains); err != nil {
		return err
	}
	if err := ValidateServiceUrl(msg.ServiceUrl); err != nil {
		return err
	}
	return nil
}

func (msg *MessageEditStakeServiceNode) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageUnstakeServiceNode) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessageUnstakeServiceNode) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageUnpauseServiceNode) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessageUnpauseServiceNode) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessagePauseServiceNode) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessagePauseServiceNode) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageStakeFisherman) ValidateBasic() types.Error {
	if err := ValidateAmount(msg.Amount); err != nil {
		return err
	}
	if err := ValidatePublicKey(msg.PublicKey); err != nil {
		return err
	}
	if err := ValidateRelayChains(msg.Chains); err != nil {
		return err
	}
	if err := ValidateServiceUrl(msg.ServiceUrl); err != nil {
		return err
	}
	return ValidateOutputAddress(msg.OutputAddress)
}

func (msg *MessageChangeParameter) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageChangeParameter) ValidateBasic() types.Error {
	if msg.ParameterKey == "" {
		return types.ErrEmptyParamKey()
	}
	if msg.ParameterValue == nil {
		return types.ErrEmptyParamValue()
	}
	if err := ValidateAddress(msg.Owner); err != nil {
		return err
	}
	return nil
}

func (msg *MessageStakeFisherman) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageEditStakeFisherman) ValidateBasic() types.Error {
	if err := ValidateAmount(msg.AmountToAdd); err != nil {
		return err
	}
	if err := ValidateAddress(msg.Address); err != nil {
		return err
	}
	if err := ValidateRelayChains(msg.Chains); err != nil {
		return err
	}
	if err := ValidateServiceUrl(msg.ServiceUrl); err != nil {
		return err
	}
	return nil
}

func (msg *MessageEditStakeFisherman) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageUnstakeFisherman) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessageUnstakeFisherman) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageUnpauseFisherman) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessageUnpauseFisherman) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessagePauseFisherman) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessagePauseFisherman) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageFishermanPauseServiceNode) ValidateBasic() types.Error {
	if err := ValidateAddress(msg.Reporter); err != nil {
		return err
	}
	return ValidateAddress(msg.Address)
}

func (msg *MessageFishermanPauseServiceNode) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageStakeValidator) ValidateBasic() types.Error {
	if err := ValidateAmount(msg.Amount); err != nil {
		return err
	}
	if err := ValidatePublicKey(msg.PublicKey); err != nil {
		return err
	}
	if err := ValidateServiceUrl(msg.ServiceUrl); err != nil {
		return err
	}
	return ValidateOutputAddress(msg.OutputAddress)
}

func (msg *MessageStakeValidator) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageEditStakeValidator) ValidateBasic() types.Error {
	// validate amount
	if err := ValidateAmount(msg.AmountToAdd); err != nil {
		return err
	}
	if err := ValidateAddress(msg.Address); err != nil {
		return err
	}
	if err := ValidateServiceUrl(msg.ServiceUrl); err != nil {
		return err
	}
	return nil
}

func (msg *MessageEditStakeValidator) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageUnstakeValidator) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessageUnstakeValidator) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageUnpauseValidator) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessageUnpauseValidator) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessagePauseValidator) ValidateBasic() types.Error {
	return ValidateAddress(msg.Address)
}

func (msg *MessagePauseValidator) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageDoubleSign) ValidateBasic() types.Error {
	if err := msg.VoteA.ValidateBasic(); err != nil {
		return err
	}
	if err := msg.VoteB.ValidateBasic(); err != nil {
		return err
	}
	if !bytes.Equal(msg.VoteA.PublicKey, msg.VoteB.PublicKey) {
		return types.ErrUnequalPublicKeys()
	}
	if msg.VoteA.Type != msg.VoteB.Type {
		return types.ErrUnequalVoteTypes()
	}
	if msg.VoteA.Height != msg.VoteB.Height {
		return types.ErrUnequalHeights()
	}
	if msg.VoteA.Round != msg.VoteB.Round {
		return types.ErrUnequalRounds()
	}
	if bytes.Equal(msg.VoteA.BlockHash, msg.VoteB.BlockHash) {
		return types.ErrEqualVotes()
	}
	return nil
}

func (msg *MessageDoubleSign) SetSigner(signer []byte) {
	msg.ReporterAddress = signer
}

func (msg *MessageSend) ValidateBasic() types.Error {
	if err := ValidateAddress(msg.FromAddress); err != nil {
		return err
	}
	if err := ValidateAddress(msg.ToAddress); err != nil {
		return err
	}
	if err := ValidateAmount(msg.Amount); err != nil {
		return err
	}
	return nil
}

func (msg *MessageSend) SetSigner(signer []byte) {
	log.Println("[NOOP] SetSigner on MessageSend")
}

func ValidateAddress(address []byte) types.Error {
	if address == nil {
		return types.ErrEmptyAddress()
	}
	addrLen := len(address)
	if addrLen != cryptoPocket.AddressLen {
		return types.ErrInvalidAddressLen(cryptoPocket.ErrInvalidAddressLen(addrLen))
	}
	return nil
}

func ValidateOutputAddress(address []byte) types.Error {
	if address == nil {
		return types.ErrNilOutputAddress()
	}
	addrLen := len(address)
	if addrLen != cryptoPocket.AddressLen {
		return types.ErrInvalidAddressLen(cryptoPocket.ErrInvalidAddressLen(addrLen))
	}
	return nil
}

func ValidatePublicKey(publicKey []byte) types.Error {
	if publicKey == nil {
		return types.ErrEmptyPublicKey()
	}
	pubKeyLen := len(publicKey)
	if pubKeyLen != cryptoPocket.PublicKeyLen {
		return types.ErrInvalidPublicKeylen(cryptoPocket.ErrInvalidPublicKeyLen(pubKeyLen))
	}
	return nil
}

func ValidateHash(hash []byte) types.Error {
	if hash == nil {
		return types.ErrEmptyHash()
	}
	hashLen := len(hash)
	if hashLen != cryptoPocket.SHA3HashLen {
		return types.ErrInvalidHashLength(cryptoPocket.ErrInvalidHashLen(hashLen))
	}
	return nil
}

func ValidateRelayChains(chains []string) types.Error {
	if chains == nil {
		return types.ErrEmptyRelayChains()
	}
	for _, chain := range chains {
		relayChain := RelayChain(chain)
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

func ValidateServiceUrl(uri string) types.Error {
	uri = strings.ToLower(uri)
	_, err := url.ParseRequestURI(uri)
	if err != nil {
		return types.ErrInvalidServiceUrl(err.Error())
	}
	if !(uri[:8] == HttpsPrefix || uri[:7] == HttpPrefix) {
		return types.ErrInvalidServiceUrl(InvalidURLPrefix)
	}
	temp := strings.Split(uri, Colon)
	if len(temp) != 3 {
		return types.ErrInvalidServiceUrl(PortRequired)
	}
	port, err := strconv.Atoi(temp[2])
	if err != nil {
		return types.ErrInvalidServiceUrl(NonNumberPort)
	}
	if port > MaxPort || port < 0 {
		return types.ErrInvalidServiceUrl(PortOutOfRange)
	}
	if !strings.Contains(uri, Period) {
		return types.ErrInvalidServiceUrl(NoPeriod)
	}
	return nil
}
