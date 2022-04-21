package types

import (
	"github.com/golang/protobuf/proto"
	shared "github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/runtime/protoiface"
)

type Marshaler interface {
	Marshal(protoiface.MessageV1) ([]byte, error)
	Unmarshal([]byte, protoiface.MessageV1) error
}

type ProtoMarshaler struct{}

func (pm *ProtoMarshaler) Marshal(m protoiface.MessageV1) ([]byte, error) {
	data, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (pm *ProtoMarshaler) Unmarshal(data []byte, msg protoiface.MessageV1) error {
	err := proto.Unmarshal(data, msg)
	if err != nil {
		msg = nil
		return err
	}
	return nil
}

func NewProtoMarshaler() *ProtoMarshaler {
	return &ProtoMarshaler{}
}

func NewP2PMessage(nonce int32, level int32, src, dest string, event *shared.PocketEvent) *P2PMessage {
	return &P2PMessage{
		Metadata: &Metadata{
			// TODO(derrandz): bring back when the hash behavior is spec'd
			//Hash:        utils.GenerateRandomHash(),
			Level:       level,
			Source:      src,
			Destination: dest,
		},
		Payload: event,
	}
}

func (msg *P2PMessage) MarkAsBroadcastMessage() {
	msg.Metadata.Broadcast = true
}
