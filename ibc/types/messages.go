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
	_ utilityTypes.Message = &UpdateIbcStore{}
	_ utilityTypes.Message = &PruneIbcStore{}
)

func (m *IbcMessage) ValidateBasic() coreTypes.Error {
	switch msg := m.Event.(type) {
	case *IbcMessage_Update:
		return msg.Update.ValidateBasic()
	case *IbcMessage_Prune:
		return msg.Prune.ValidateBasic()
	default:
		return coreTypes.ErrUnknownIBCMessageType(fmt.Sprintf("%T", msg))
	}
}

func (m *UpdateIbcStore) ValidateBasic() coreTypes.Error {
	if m.Key == nil {
		return coreTypes.ErrNilField("key")
	}
	if m.Prefix == nil {
		return coreTypes.ErrNilField("prefix")
	}
	if m.Value == nil {
		return coreTypes.ErrNilField("value")
	}
	return nil
}

func (m *PruneIbcStore) ValidateBasic() coreTypes.Error {
	if m.Key == nil {
		return coreTypes.ErrNilField("key")
	}
	if m.Prefix == nil {
		return coreTypes.ErrNilField("prefix")
	}
	return nil
}

func (m *UpdateIbcStore) SetSigner(signer []byte) { m.Signer = signer }
func (m *PruneIbcStore) SetSigner(signer []byte)  { m.Signer = signer }

func (m *UpdateIbcStore) GetMessageName() string {
	return string(m.ProtoReflect().Descriptor().Name())
}

func (m *PruneIbcStore) GetMessageName() string {
	return string(m.ProtoReflect().Descriptor().Name())
}

func (m *UpdateIbcStore) GetMessageRecipient() string { return "" }
func (m *PruneIbcStore) GetMessageRecipient() string  { return "" }

func (m *UpdateIbcStore) GetSigner() []byte { return m.Signer }
func (m *PruneIbcStore) GetSigner() []byte  { return m.Signer }

func (m *UpdateIbcStore) GetActorType() coreTypes.ActorType {
	return coreTypes.ActorType_ACTOR_TYPE_VAL
}

func (m *PruneIbcStore) GetActorType() coreTypes.ActorType {
	return coreTypes.ActorType_ACTOR_TYPE_VAL
}

func (m *UpdateIbcStore) GetCanonicalBytes() []byte {
	bz, err := codec.GetCodec().Marshal(m)
	if err != nil {
		log.Fatalf("must marshal %v", err)
	}
	return bz // DISCUSS(#142): should we also sort the JSON like in V0?
}

func (m *PruneIbcStore) GetCanonicalBytes() []byte {
	bz, err := codec.GetCodec().Marshal(m)
	if err != nil {
		log.Fatalf("must marshal %v", err)
	}
	return bz // DISCUSS(#142): should we also sort the JSON like in V0?
}
