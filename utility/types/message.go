package types

import (
	"bytes"
	"net/url"
	"strconv"
	"strings"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/proto"
)

// message.go contains ValidateBasic() and SetSigner logic for all message types
// ValidateBasic() is a 'stateless' validation check that should encapsulate all
// validations possible before even checking the state storage layer

const (
	UnstakingStatus = 1
	StakedStatus    = 2
)

const (
	MillionInt       = 1000000
	ZeroInt          = 0
	HeightNotUsed    = -1
	EmptyString      = ""
	HttpsPrefix      = "https://"
	HttpPrefix       = "http://"
	Colon            = ":"
	Period           = "."
	InvalidURLPrefix = "the url must start with http:// or https://"
	PortRequired     = "a port is required"
	NonNumberPort    = "invalid port, cant convert to integer"
	PortOutOfRange   = "invalid port, out of valid port range"
	NoPeriod         = "must contain one '.'"
	MaxPort          = 65535
)

type Message interface {
	proto.Message

	SetSigner(signer []byte)
	ValidateBasic() types.Error
	GetActorType() ActorType
}

func (msg *MessageStake) ValidateBasic() types.Error {
	if err := ValidatePublicKey(msg.GetPublicKey()); err != nil {
		return err
	}
	if err := ValidateOutputAddress(msg.GetOutputAddress()); err != nil {
		return err
	}
	return ValidateStaker(msg)
}

func (msg *MessageEditStake) ValidateBasic() types.Error {
	if err := ValidateAddress(msg.GetAddress()); err != nil {
		return err
	}
	return ValidateStaker(msg)
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

func (msg *MessageUnstake) ValidateBasic() types.Error      { return ValidateAddress(msg.Address) }
func (msg *MessageUnpause) ValidateBasic() types.Error      { return ValidateAddress(msg.Address) }
func (msg *MessageStake) SetSigner(signer []byte)           { msg.Signer = signer }
func (msg *MessageEditStake) SetSigner(signer []byte)       { msg.Signer = signer }
func (msg *MessageUnstake) SetSigner(signer []byte)         { msg.Signer = signer }
func (msg *MessageUnpause) SetSigner(signer []byte)         { msg.Signer = signer }
func (msg *MessageDoubleSign) SetSigner(signer []byte)      { msg.ReporterAddress = signer }
func (msg *MessageSend) SetSigner(signer []byte)            { /*no op*/ }
func (msg *MessageChangeParameter) SetSigner(signer []byte) { msg.Signer = signer }
func (x *MessageChangeParameter) GetActorType() ActorType   { return -1 }
func (x *MessageDoubleSign) GetActorType() ActorType        { return -1 }

// helpers

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

func ValidateActorType(_ ActorType) types.Error {
	// TODO (team) not sure if there's anything we can do here
	return nil
}

func ValidateServiceUrl(actorType ActorType, uri string) types.Error {
	if actorType == ActorType_App {
		return nil
	}
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

const (
	RelayChainLength = 4 // pre-determined length that strikes a balance between combination possibilities & storage
)

type RelayChain string

// TODO: Consider adding a governance parameter for a list of valid relay chains
func (rc *RelayChain) Validate() types.Error {
	if rc == nil || *rc == "" {
		return types.ErrEmptyRelayChain()
	}
	rcLen := len(*rc)
	if rcLen != RelayChainLength {
		return types.ErrInvalidRelayChainLength(rcLen, RelayChainLength)
	}
	return nil
}

func ValidateStaker(msg MessageStaker) types.Error {
	if err := ValidateActorType(msg.GetActorType()); err != nil {
		return err
	}
	if err := ValidateAmount(msg.GetAmount()); err != nil {
		return err
	}
	if err := ValidateRelayChains(msg.GetChains()); err != nil {
		return err
	}
	return ValidateServiceUrl(msg.GetActorType(), msg.GetServiceUrl())
}

type MessageStaker interface {
	GetActorType() ActorType
	GetAmount() string
	GetChains() []string
	GetServiceUrl() string
}
