package types

import (
	"bytes"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/proto"
)

/*
`message.go`` contains `ValidateBasic` and `SetSigner`` logic for all message types.

`ValidateBasic` is a **stateless** validation check that should encapsulate all
validations possible before even checking the state storage layer.
*/

// CLEANUP: Move these to a better shared location or inline the vars.
const (
	MillionInt       = 1000000
	ZeroInt          = 0
	HeightNotUsed    = int64(-1)
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
	ValidateBasic() Error
	GetActorType() UtilActorType

	// Get the canonical byte representation of the ProtoMsg.
	GetSignBytes() []byte
}

func (msg *MessageStake) ValidateBasic() Error {
	if err := ValidatePublicKey(msg.GetPublicKey()); err != nil {
		return err
	}
	if err := ValidateOutputAddress(msg.GetOutputAddress()); err != nil {
		return err
	}
	return ValidateStaker(msg)
}

func (msg *MessageEditStake) ValidateBasic() Error {
	if err := ValidateAddress(msg.GetAddress()); err != nil {
		return err
	}
	return ValidateStaker(msg)
}

func (msg *MessageDoubleSign) ValidateBasic() Error {
	if err := msg.VoteA.ValidateBasic(); err != nil {
		return err
	}
	if err := msg.VoteB.ValidateBasic(); err != nil {
		return err
	}
	if !bytes.Equal(msg.VoteA.PublicKey, msg.VoteB.PublicKey) {
		return ErrUnequalPublicKeys()
	}
	if msg.VoteA.Type != msg.VoteB.Type {
		return ErrUnequalVoteTypes()
	}
	if msg.VoteA.Height != msg.VoteB.Height {
		return ErrUnequalHeights()
	}
	if msg.VoteA.Round != msg.VoteB.Round {
		return ErrUnequalRounds()
	}
	if bytes.Equal(msg.VoteA.BlockHash, msg.VoteB.BlockHash) {
		return ErrEqualVotes()
	}
	return nil
}

func (msg *MessageSend) ValidateBasic() Error {
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

func (msg *MessageChangeParameter) ValidateBasic() Error {
	if msg.ParameterKey == "" {
		return ErrEmptyParamKey()
	}
	if msg.ParameterValue == nil {
		return ErrEmptyParamValue()
	}
	if err := ValidateAddress(msg.Owner); err != nil {
		return err
	}
	return nil
}

func (msg *MessageUnstake) ValidateBasic() Error              { return ValidateAddress(msg.Address) }
func (msg *MessageUnpause) ValidateBasic() Error              { return ValidateAddress(msg.Address) }
func (msg *MessageStake) SetSigner(signer []byte)             { msg.Signer = signer }
func (msg *MessageEditStake) SetSigner(signer []byte)         { msg.Signer = signer }
func (msg *MessageUnstake) SetSigner(signer []byte)           { msg.Signer = signer }
func (msg *MessageUnpause) SetSigner(signer []byte)           { msg.Signer = signer }
func (msg *MessageDoubleSign) SetSigner(signer []byte)        { msg.ReporterAddress = signer }
func (msg *MessageSend) SetSigner(signer []byte)              { /*no op*/ }
func (msg *MessageChangeParameter) SetSigner(signer []byte)   { msg.Signer = signer }
func (x *MessageChangeParameter) GetActorType() UtilActorType { return -1 }
func (x *MessageDoubleSign) GetActorType() UtilActorType      { return -1 }
func (msg *MessageStake) GetSignBytes() []byte                { return getSignBytes(msg) }
func (msg *MessageEditStake) GetSignBytes() []byte            { return getSignBytes(msg) }
func (msg *MessageDoubleSign) GetSignBytes() []byte           { return getSignBytes(msg) }
func (msg *MessageSend) GetSignBytes() []byte                 { return getSignBytes(msg) }
func (msg *MessageChangeParameter) GetSignBytes() []byte      { return getSignBytes(msg) }
func (msg *MessageUnstake) GetSignBytes() []byte              { return getSignBytes(msg) }
func (msg *MessageUnpause) GetSignBytes() []byte              { return getSignBytes(msg) }

// helpers

func ValidateAddress(address []byte) Error {
	if address == nil {
		return ErrEmptyAddress()
	}
	addrLen := len(address)
	if addrLen != cryptoPocket.AddressLen {
		return ErrInvalidAddressLen(cryptoPocket.ErrInvalidAddressLen(addrLen))
	}
	return nil
}

func ValidateOutputAddress(address []byte) Error {
	if address == nil {
		return ErrNilOutputAddress()
	}
	addrLen := len(address)
	if addrLen != cryptoPocket.AddressLen {
		return ErrInvalidAddressLen(cryptoPocket.ErrInvalidAddressLen(addrLen))
	}
	return nil
}

func ValidatePublicKey(publicKey []byte) Error {
	if publicKey == nil {
		return ErrEmptyPublicKey()
	}
	pubKeyLen := len(publicKey)
	if pubKeyLen != cryptoPocket.PublicKeyLen {
		return ErrInvalidPublicKeylen(cryptoPocket.ErrInvalidPublicKeyLen(pubKeyLen))
	}
	return nil
}

func ValidateHash(hash []byte) Error {
	if hash == nil {
		return ErrEmptyHash()
	}
	hashLen := len(hash)
	if hashLen != cryptoPocket.SHA3HashLen {
		return ErrInvalidHashLength(cryptoPocket.ErrInvalidHashLen(hashLen))
	}
	return nil
}

func ValidateRelayChains(chains []string) Error {
	if chains == nil {
		return ErrEmptyRelayChains()
	}
	for _, chain := range chains {
		relayChain := RelayChain(chain)
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

func ValidateActorType(_ UtilActorType) Error {
	// TODO (team) not sure if there's anything we can do here
	return nil
}

func ValidateServiceUrl(actorType UtilActorType, uri string) Error {
	if actorType == UtilActorType_App {
		return nil
	}
	uri = strings.ToLower(uri)
	_, err := url.ParseRequestURI(uri)
	if err != nil {
		return ErrInvalidServiceUrl(err.Error())
	}
	if !(uri[:8] == HttpsPrefix || uri[:7] == HttpPrefix) {
		return ErrInvalidServiceUrl(InvalidURLPrefix)
	}
	temp := strings.Split(uri, Colon)
	if len(temp) != 3 {
		return ErrInvalidServiceUrl(PortRequired)
	}
	port, err := strconv.Atoi(temp[2])
	if err != nil {
		return ErrInvalidServiceUrl(NonNumberPort)
	}
	if port > MaxPort || port < 0 {
		return ErrInvalidServiceUrl(PortOutOfRange)
	}
	if !strings.Contains(uri, Period) {
		return ErrInvalidServiceUrl(NoPeriod)
	}
	return nil
}

// CLEANUP: Figure out where these other types should be defined.
//          It's a bit weird that they are hidden at the bottom of the file.

const (
	RelayChainLength = 4 // pre-determined length that strikes a balance between combination possibilities & storage
)

type RelayChain string

// TODO: Consider adding a governance parameter for a list of valid relay chains
func (rc *RelayChain) Validate() Error {
	if rc == nil || *rc == "" {
		return ErrEmptyRelayChain()
	}
	rcLen := len(*rc)
	if rcLen != RelayChainLength {
		return ErrInvalidRelayChainLength(rcLen, RelayChainLength)
	}
	return nil
}

type MessageStaker interface {
	GetActorType() UtilActorType
	GetAmount() string
	GetChains() []string
	GetServiceUrl() string
}

func ValidateStaker(msg MessageStaker) Error {
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

func getSignBytes(msg Message) []byte {
	bz, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		log.Fatalf("must marshal %v", err)
	}
	// DISCUSS(team): should we also sort the JSON like in V0?
	return bz
}
