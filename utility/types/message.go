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
	proto.Message
	Validatable

	SetSigner(signer []byte)
	GetCanonicalBytes() []byte
	GetActorType() coreTypes.ActorType
	GetMessageName() string
	GetMessageRecipient() string
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
	if err := ValidatePublicKey(msg.GetPublicKey()); err != nil {
		return err
	}
	if err := ValidateOutputAddress(msg.GetOutputAddress()); err != nil {
		return err
	}
	return ValidateStaker(msg)
}

func (msg *MessageEditStake) ValidateBasic() Error {
	if err := validateAddress(msg.GetAddress()); err != nil {
		return err
	}
	return ValidateStaker(msg)
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
func (msg *MessageUnstake) GetMessageName() string         { return getMessageType(msg) }
func (msg *MessageUnpause) GetMessageName() string         { return getMessageType(msg) }
func (msg *MessageEditStake) GetMessageName() string       { return getMessageType(msg) }
func (msg *MessageStake) GetMessageName() string           { return getMessageType(msg) }
func (msg *MessageChangeParameter) GetMessageName() string { return getMessageType(msg) }

func (msg *MessageSend) GetMessageRecipient() string            { return hex.EncodeToString(msg.ToAddress) }
func (msg *MessageUnstake) GetMessageRecipient() string         { return "" }
func (msg *MessageUnpause) GetMessageRecipient() string         { return "" }
func (msg *MessageEditStake) GetMessageRecipient() string       { return "" }
func (msg *MessageStake) GetMessageRecipient() string           { return "" }
func (msg *MessageChangeParameter) GetMessageRecipient() string { return "" }

func (msg *MessageUnstake) ValidateBasic() Error { return validateAddress(msg.Address) }
func (msg *MessageUnpause) ValidateBasic() Error { return validateAddress(msg.Address) }

func (msg *MessageStake) SetSigner(signer []byte)           { msg.Signer = signer }
func (msg *MessageEditStake) SetSigner(signer []byte)       { msg.Signer = signer }
func (msg *MessageUnstake) SetSigner(signer []byte)         { msg.Signer = signer }
func (msg *MessageUnpause) SetSigner(signer []byte)         { msg.Signer = signer }
func (msg *MessageSend) SetSigner(signer []byte)            { /*no op*/ }
func (msg *MessageChangeParameter) SetSigner(signer []byte) { msg.Signer = signer }

func (msg *MessageSend) GetActorType() coreTypes.ActorType {
	return coreTypes.ActorType_ACTOR_TYPE_UNSPECIFIED // there's no actor type for message send, so return zero to allow fee retrieval
}
func (x *MessageChangeParameter) GetActorType() coreTypes.ActorType { return -1 }

func (msg *MessageStake) GetCanonicalBytes() []byte           { return getCanonicalBytes(msg) }
func (msg *MessageEditStake) GetCanonicalBytes() []byte       { return getCanonicalBytes(msg) }
func (msg *MessageSend) GetCanonicalBytes() []byte            { return getCanonicalBytes(msg) }
func (msg *MessageChangeParameter) GetCanonicalBytes() []byte { return getCanonicalBytes(msg) }
func (msg *MessageUnstake) GetCanonicalBytes() []byte         { return getCanonicalBytes(msg) }
func (msg *MessageUnpause) GetCanonicalBytes() []byte         { return getCanonicalBytes(msg) }

// helpers

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
		return ErrInvalidPublicKeyLen(cryptoPocket.ErrInvalidPublicKeyLen(pubKeyLen))
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
		relayChain := relayChain(chain)
		if err := relayChain.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

func validateAmount(amount string) Error {
	if amount == "" {
		return ErrEmptyAmount()
	}
	if _, err := converters.StringToBigInt(amount); err != nil {
		return ErrStringToBigInt()
	}
	return nil
}

func ValidateActorType(_ coreTypes.ActorType) Error {
	// TODO (team) not sure if there's anything we can do here
	return nil
}

func ValidateServiceUrl(actorType coreTypes.ActorType, uri string) Error {
	if actorType == coreTypes.ActorType_ACTOR_TYPE_APP {
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

func getMessageType(msg Message) string {
	return string(msg.ProtoReflect().Descriptor().Name())
}

type messageStaker interface {
	GetActorType() coreTypes.ActorType
	GetAmount() string
	GetChains() []string
	GetServiceUrl() string
}

func ValidateStaker(msg messageStaker) Error {
	if err := ValidateActorType(msg.GetActorType()); err != nil {
		return err
	}
	if err := validateAmount(msg.GetAmount()); err != nil {
		return err
	}
	if err := ValidateRelayChains(msg.GetChains()); err != nil {
		return err
	}
	return ValidateServiceUrl(msg.GetActorType(), msg.GetServiceUrl())
}

func getCanonicalBytes(msg Message) []byte {
	bz, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		log.Fatalf("must marshal %v", err)
	}
	// DISCUSS(#142): should we also sort the JSON like in V0?
	return bz
}
