package types

import (
	"log"

	"github.com/blang/semver/v4"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

var (
	_ Message = &MessageUpgrade{}
)

func (msg *MessageUpgrade) ValidateBasic() coreTypes.Error {
	if err := validateAddress(msg.Signer); err != nil {
		return err
	}
	if _, err := semver.Parse(msg.Version); err != nil {
		return coreTypes.ErrInvalidProtocolVersion(msg.Version)
	}
	if msg.Height < 1 {
		return coreTypes.ErrInvalidBlockHeight()
	}
	return nil
}

func (msg *MessageUpgrade) SetSigner(signer []byte) {
	msg.Signer = signer
}

func (msg *MessageUpgrade) GetMessageName() string {
	return getMessageType(msg)
}

func (msg *MessageUpgrade) GetMessageRecipient() string {
	// Upgrade message does not have a recipient
	return ""
}

func (msg *MessageUpgrade) GetActorType() coreTypes.ActorType {
	return coreTypes.ActorType_ACTOR_TYPE_UNSPECIFIED
}

func (msg *MessageUpgrade) GetCanonicalBytes() []byte {
	bz, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		log.Fatalf("must marshal %v", err)
	}
	return bz
}
