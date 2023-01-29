package types

import (
	"encoding/hex"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/proto"
)

// A message is a component of a transaction (excluding metadata such as the signature)
// defining the action driving the state transition
type Message interface {
	proto.Message // TECHDEBT: Should not be refere
	Validatable

	SetSigner(signer []byte)
	GetActorType() coreTypes.ActorType
	GetMessageName() string
	GetMessageRecipient() string
	GetCanonicalBytes() []byte
}

var (
	_ Message = &MessageSend{}
	_ Message = &MessageStake{}
	_ Message = &MessageEditStake{}
	_ Message = &MessageUnstake{}
	_ Message = &MessageUnpause{}
	_ Message = &MessageChangeParameter{}
)

func (msg *MessageSend) ValidateBasic() Error {
	if err := validateAddress(msg.FromAddress); err != nil {
		return err
	}
	if err := validateAddress(msg.ToAddress); err != nil {
		return err
	}
	if err := validateAmount(msg.Amount); err != nil {
		return err
	}
	return nil
}
func (msg *MessageStake) ValidateBasic() Error {
	if err := validatePublicKey(msg.PublicKey); err != nil {
		return err
	}
	if err := validateOutputAddress(msg.OutputAddress); err != nil {
		return err
	}
	return validateStaker(msg)
}
func (msg *MessageUnstake) ValidateBasic() Error {
	return validateAddress(msg.Address)
}
func (msg *MessageUnpause) ValidateBasic() Error {
	return validateAddress(msg.Address)
}
func (msg *MessageEditStake) ValidateBasic() Error {
	if err := validateAddress(msg.Address); err != nil {
		return err
	}
	return validateStaker(msg)
}
func (msg *MessageChangeParameter) ValidateBasic() Error {
	if msg.ParameterKey == "" {
		return ErrEmptyParamKey()
	}
	if msg.ParameterValue == nil {
		return ErrEmptyParamValue()
	}
	if err := validateAddress(msg.Owner); err != nil {
		return err
	}
	return nil
}

func (msg *MessageSend) GetMessageName() string            { return getMessageType(msg) }
func (msg *MessageStake) GetMessageName() string           { return getMessageType(msg) }
func (msg *MessageEditStake) GetMessageName() string       { return getMessageType(msg) }
func (msg *MessageUnstake) GetMessageName() string         { return getMessageType(msg) }
func (msg *MessageUnpause) GetMessageName() string         { return getMessageType(msg) }
func (msg *MessageChangeParameter) GetMessageName() string { return getMessageType(msg) }

func (msg *MessageSend) GetMessageRecipient() string            { return hex.EncodeToString(msg.ToAddress) }
func (msg *MessageStake) GetMessageRecipient() string           { return "" }
func (msg *MessageEditStake) GetMessageRecipient() string       { return "" }
func (msg *MessageUnstake) GetMessageRecipient() string         { return "" }
func (msg *MessageUnpause) GetMessageRecipient() string         { return "" }
func (msg *MessageChangeParameter) GetMessageRecipient() string { return "" }

func (msg *MessageSend) SetSigner(signer []byte)            { /*no op*/ }
func (msg *MessageStake) SetSigner(signer []byte)           { msg.Signer = signer }
func (msg *MessageEditStake) SetSigner(signer []byte)       { msg.Signer = signer }
func (msg *MessageUnstake) SetSigner(signer []byte)         { msg.Signer = signer }
func (msg *MessageUnpause) SetSigner(signer []byte)         { msg.Signer = signer }
func (msg *MessageChangeParameter) SetSigner(signer []byte) { msg.Signer = signer }

func (msg *MessageSend) GetActorType() coreTypes.ActorType {
	return coreTypes.ActorType_ACTOR_TYPE_UNSPECIFIED // there's no actor type for message send, so return zero to allow fee retrieval
}
func (msg *MessageChangeParameter) GetActorType() coreTypes.ActorType {
	return -1 // CONSIDERATION: Should we create an actor for the DAO or ACLed addresses?
}

func (msg *MessageSend) GetCanonicalBytes() []byte            { return getCanonicalBytes(msg) }
func (msg *MessageStake) GetCanonicalBytes() []byte           { return getCanonicalBytes(msg) }
func (msg *MessageEditStake) GetCanonicalBytes() []byte       { return getCanonicalBytes(msg) }
func (msg *MessageUnstake) GetCanonicalBytes() []byte         { return getCanonicalBytes(msg) }
func (msg *MessageUnpause) GetCanonicalBytes() []byte         { return getCanonicalBytes(msg) }
func (msg *MessageChangeParameter) GetCanonicalBytes() []byte { return getCanonicalBytes(msg) }

// Helpers

// CONSIDERATION: If the protobufs contain semantic types, we could potentially leverage
//                a shared `address.ValidateBasic()` throughout the codebase.s
func validateAddress(address []byte) Error {
	if address == nil {
		return ErrEmptyAddress()
	}
	addrLen := len(address)
	if addrLen != cryptoPocket.AddressLen {
		return ErrInvalidAddressLen(cryptoPocket.ErrInvalidAddressLen(addrLen))
	}
	return nil
}

// CONSIDERATION: Consolidate with `validateAddress`?
func validateOutputAddress(address []byte) Error {
	if address == nil {
		return ErrNilOutputAddress()
	}
	addrLen := len(address)
	if addrLen != cryptoPocket.AddressLen {
		return ErrInvalidAddressLen(cryptoPocket.ErrInvalidAddressLen(addrLen))
	}
	return nil
}

// CONSIDERATION: If the protobufs contain semantic types, we could potentially leverage
//                a shared `address.ValidateBasic()` throughout the codebase.s
func validatePublicKey(publicKey []byte) Error {
	if publicKey == nil {
		return ErrEmptyPublicKey()
	}
	pubKeyLen := len(publicKey)
	if pubKeyLen != cryptoPocket.PublicKeyLen {
		return ErrInvalidPublicKeyLen(pubKeyLen)
	}
	return nil
}

func ValidateHash(hash []byte) Error {
	if hash == nil {
		return ErrEmptyHash()
	}
	hashLen := len(hash)
	if hashLen != cryptoPocket.SHA3HashLen {
		return ErrInvalidHashLength(hashLen)
	}
	return nil
}

func validateRelayChains(chains []string) Error {
	if chains == nil {
		return ErrEmptyRelayChains()
	}
	for _, chain := range chains {
		if err := relayChain(chain).ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

func getMessageType(msg Message) string {
	return string(msg.ProtoReflect().Descriptor().Name())
}

// An internal helper interface to consolidate logic related to validating staking related messages
type stakingMessage interface {
	GetActorType() coreTypes.ActorType
	GetAmount() string
	GetChains() []string
	GetServiceUrl() string
}

func validateStaker(msg stakingMessage) Error {
	if err := validateActorType(msg.GetActorType()); err != nil {
		return err
	}
	if err := validateAmount(msg.GetAmount()); err != nil {
		return err
	}
	if err := validateRelayChains(msg.GetChains()); err != nil {
		return err
	}
	return validateServiceUrl(msg.GetActorType(), msg.GetServiceUrl())
}

func getCanonicalBytes(msg Message) []byte {
	bz, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		log.Fatalf("must marshal %v", err)
	}
	// DISCUSS(#142): should we also sort the JSON like in V0?
	return bz
}

func validateActorType(_ coreTypes.ActorType) Error {
	// TODO: Is there any sort of validation that should be done here?
	return nil
}

func validateAmount(amount string) Error {
	if amount == "" {
		return ErrEmptyAmount()
	}
	if _, err := converters.StringToBigInt(amount); err != nil {
		return ErrStringToBigInt(err)
	}
	return nil
}

func validateServiceUrl(actorType coreTypes.ActorType, uri string) Error {
	if actorType == coreTypes.ActorType_ACTOR_TYPE_APP {
		return nil
	}

	uri = strings.ToLower(uri)
	_, err := url.ParseRequestURI(uri)
	if err != nil {
		return ErrInvalidServiceUrl(err.Error())
	}
	if !(uri[:8] == httpsPrefix || uri[:7] == httpPrefix) {
		return ErrInvalidServiceUrl(invalidURLPrefix)
	}

	urlParts := strings.Split(uri, colon)
	if len(urlParts) != 3 { // protocol:host:port
		return ErrInvalidServiceUrl(portRequired)
	}
	port, err := strconv.Atoi(urlParts[2])
	if err != nil {
		return ErrInvalidServiceUrl(NonNumberPort)
	}
	if port > maxPort || port < 0 {
		return ErrInvalidServiceUrl(PortOutOfRange)
	}
	if !strings.Contains(uri, period) {
		return ErrInvalidServiceUrl(NoPeriod)
	}
	return nil
}
