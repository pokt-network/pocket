package types

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	utilityTypes "github.com/pokt-network/pocket/utility/types"
)

// Implement the Message interface
var (
	_ utilityTypes.Message = &UpdateIBCStore{}
	_ utilityTypes.Message = &PruneIBCStore{}
)

func (m *IBCMessage) ValidateBasic() coreTypes.Error {
	switch msg := m.Event.(type) {
	case *IBCMessage_Update:
		return msg.Update.ValidateBasic()
	case *IBCMessage_Prune:
		return msg.Prune.ValidateBasic()
	default:
		return coreTypes.ErrUnknownIBCMessageType(fmt.Sprintf("%T", msg))
	}
}

func (m *UpdateIBCStore) ValidateBasic() coreTypes.Error {
	if m.Key == nil {
		return coreTypes.ErrNilField("key")
	}
	if m.Value == nil {
		return coreTypes.ErrNilField("value")
	}
	return nil
}

func (m *PruneIBCStore) ValidateBasic() coreTypes.Error {
	if m.Key == nil {
		return coreTypes.ErrNilField("key")
	}
	return nil
}

func (m *UpdateIBCStore) SetSigner(signer []byte) { m.Signer = signer }
func (m *PruneIBCStore) SetSigner(signer []byte)  { m.Signer = signer }

func (m *UpdateIBCStore) GetMessageName() string {
	return string(m.ProtoReflect().Descriptor().Name())
}

func (m *PruneIBCStore) GetMessageName() string {
	return string(m.ProtoReflect().Descriptor().Name())
}

func (m *UpdateIBCStore) GetMessageRecipient() string { return "" }
func (m *PruneIBCStore) GetMessageRecipient() string  { return "" }

func (m *UpdateIBCStore) GetActorType() coreTypes.ActorType {
	return coreTypes.ActorType_ACTOR_TYPE_VAL
}

func (m *PruneIBCStore) GetActorType() coreTypes.ActorType {
	return coreTypes.ActorType_ACTOR_TYPE_VAL
}

func (m *UpdateIBCStore) GetCanonicalBytes() []byte {
	bz, err := codec.GetCodec().Marshal(m)
	if err != nil {
		log.Fatalf("must marshal %v", err)
	}
	return bz // DISCUSS(#142): should we also sort the JSON like in V0?
}

func (m *PruneIBCStore) GetCanonicalBytes() []byte {
	bz, err := codec.GetCodec().Marshal(m)
	if err != nil {
		log.Fatalf("must marshal %v", err)
	}
	return bz // DISCUSS(#142): should we also sort the JSON like in V0?
}
