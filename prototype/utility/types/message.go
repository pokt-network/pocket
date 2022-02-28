package types

import (
	"bytes"
	"google.golang.org/protobuf/proto"
	"net/url"
	crypto2 "pocket/shared/crypto"
	"strconv"
	"strings"
)

type Message interface {
	proto.Message
	SetSigner(signer []byte)
	ValidateBasic() Error
}

var (
	_ Message = &MessageStakeApp{}
)

func (x *MessageStakeApp) ValidateBasic() Error {
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
	return ValidateAddress(x.OutputAddress)
}

func (x *MessageStakeApp) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageEditStakeApp) ValidateBasic() Error {
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

func (x *MessageUnstakeApp) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessageUnstakeApp) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnpauseApp) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessageUnpauseApp) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessagePauseApp) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessagePauseApp) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageStakeServiceNode) ValidateBasic() Error {
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
	if err := ValidateServiceURL(x.ServiceURL); err != nil {
		return err
	}
	// validate output address
	return ValidateAddress(x.OutputAddress)
}

func (x *MessageStakeServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageEditStakeServiceNode) ValidateBasic() Error {
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
	if err := ValidateServiceURL(x.ServiceURL); err != nil {
		return err
	}
	return nil
}

func (x *MessageEditStakeServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnstakeServiceNode) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessageUnstakeServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnpauseServiceNode) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessageUnpauseServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessagePauseServiceNode) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessagePauseServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageStakeFisherman) ValidateBasic() Error {
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
	if err := ValidateServiceURL(x.ServiceURL); err != nil {
		return err
	}
	// validate output address
	return ValidateAddress(x.OutputAddress)
}

func (x *MessageChangeParameter) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageChangeParameter) ValidateBasic() Error {
	if x.ParameterKey == "" {
		return ErrEmptyParamKey()
	}
	if x.ParameterValue == nil {
		return ErrEmptyParamValue()
	}
	if err := ValidateAddress(x.Owner); err != nil {
		return err
	}
	return nil
}

func (x *MessageStakeFisherman) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageEditStakeFisherman) ValidateBasic() Error {
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
	if err := ValidateServiceURL(x.ServiceURL); err != nil {
		return err
	}
	return nil
}

func (x *MessageEditStakeFisherman) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnstakeFisherman) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessageUnstakeFisherman) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnpauseFisherman) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessageUnpauseFisherman) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessagePauseFisherman) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessagePauseFisherman) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageFishermanPauseServiceNode) ValidateBasic() Error {
	if err := ValidateAddress(x.Reporter); err != nil {
		return err
	}
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessageFishermanPauseServiceNode) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageStakeValidator) ValidateBasic() Error {
	// validate amount
	if err := ValidateAmount(x.Amount); err != nil {
		return err
	}
	if err := ValidatePublicKey(x.PublicKey); err != nil {
		return err
	}
	if err := ValidateServiceURL(x.ServiceURL); err != nil {
		return err
	}
	// validate output address
	return ValidateAddress(x.OutputAddress)
}

func (x *MessageStakeValidator) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageEditStakeValidator) ValidateBasic() Error {
	// validate amount
	if err := ValidateAmount(x.AmountToAdd); err != nil {
		return err
	}
	if err := ValidateAddress(x.Address); err != nil {
		return err
	}
	if err := ValidateServiceURL(x.ServiceURL); err != nil {
		return err
	}
	return nil
}

func (x *MessageEditStakeValidator) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnstakeValidator) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessageUnstakeValidator) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageUnpauseValidator) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessageUnpauseValidator) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessagePauseValidator) ValidateBasic() Error {
	return ValidateAddressAndSigner(x.Address, x.Signer)
}

func (x *MessagePauseValidator) SetSigner(signer []byte) {
	x.Signer = signer
}

func (x *MessageDoubleSign) ValidateBasic() Error {
	if err := x.VoteA.ValidateBasic(); err != nil {
		return err
	}
	if err := x.VoteB.ValidateBasic(); err != nil {
		return err
	}
	if !bytes.Equal(x.VoteA.PublicKey, x.VoteB.PublicKey) {
		return ErrUnequalPublicKeys()
	}
	if x.VoteA.Type != x.VoteB.Type {
		return ErrUnequalVoteTypes()
	}
	if x.VoteA.Round != x.VoteB.Round {
		return ErrUnequalRounds()
	}
	if bytes.Equal(x.VoteA.BlockHash, x.VoteB.BlockHash) {
		return ErrEqualVotes()
	}
	return nil
}

func (x *MessageDoubleSign) SetSigner(signer []byte) {
	x.ReporterAddress = signer
}

func (x *MessageSend) ValidateBasic() Error {
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

func ValidateAddressAndSigner(address, signer []byte) Error {
	if err := ValidateAddress(address); err != nil {
		return err
	}
	return ValidateAddress(signer)
}

func ValidateAddress(address []byte) Error {
	if address == nil {
		return ErrNilOutputAddress()
	}
	if len(address) != crypto2.AddressLen {
		return ErrInvalidAddressLen(crypto2.ErrInvalidAddressLen())
	}
	return nil
}

func ValidatePublicKey(publicKey []byte) Error {
	// validate public key
	if publicKey == nil {
		return ErrEmptyPublicKey()
	}
	if len(publicKey) != crypto2.PublicKeyLen {
		return ErrInvalidPublicKeylen(crypto2.ErrInvalidPublicKeyLen())
	}
	return nil
}

func ValidateHash(hash []byte) Error {
	if hash == nil {
		return ErrEmptyHash()
	}
	if len(hash) != crypto2.SHA3HashLen {
		return ErrInvalidHashLength(crypto2.ErrInvalidHashLen())
	}
	return nil
}

func ValidateRelayChains(chains []string) Error {
	if chains == nil {
		return ErrEmptyRelayChains()
	}
	for i := 0; i < len(chains); i++ {
		relayChain := RelayChain(chains[i])
		if err := relayChain.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func ValidateAmount(amount string) Error {
	if amount == "" {
		return ErrEmptyAmount()
	}
	if _, err := StringToBigInt(amount); err != nil {
		return err
	}
	return nil
}

func ValidateServiceURL(u string) Error {
	u = strings.ToLower(u)
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return ErrInvalidServiceURL(err.Error())
	}
	if u[:8] != HttpsPrefix && u[:7] != HttpPrefix {
		return ErrInvalidServiceURL(InvalidURLPrefix)
	}
	temp := strings.Split(u, Colon)
	if len(temp) != 3 {
		return ErrInvalidServiceURL(PortRequired)
	}
	port, err := strconv.Atoi(temp[2])
	if err != nil {
		return ErrInvalidServiceURL(NonNumberPort)
	}
	if port > 65535 || port < 0 {
		return ErrInvalidServiceURL(PortOutOfRange)
	}
	if !strings.Contains(u, Period) {
		return ErrInvalidServiceURL(NoPeriod)
	}
	return nil
}
