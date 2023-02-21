package types

import (
	"encoding/hex"
	"log"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/proto"
)

// A message is a component of a transaction (excluding metadata such as the signature)
// defining the action driving the state transition

type Message interface {
	proto.Message // TECHDEBT: Still making direct `proto` reference even with a central `codec` package
	Validatable

	GetMessageName() string
	GetMessageRecipient() string
	SetSigner(signer []byte)
	GetActorType() coreTypes.ActorType
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

// CONSIDERATION: If the protobufs contain semantic types (e.g. Address is an interface), we could
// potentially leverage a shared `address.ValidateBasic()` throughout the codebase.
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

// CONSIDERATION: Consolidate with `validateAddress`? The only difference is the error message.
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
//
//	a shared `address.ValidateBasic()` throughout the codebase.s
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

//nolint:unused // TODO: need to figure out why this function was added and never used
func validateHash(hash []byte) Error {
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

func getCanonicalBytes(msg Message) []byte {
	bz, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		log.Fatalf("must marshal %v", err)
	}
	return bz // DISCUSS(#142): should we also sort the JSON like in V0?
}
