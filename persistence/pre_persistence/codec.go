package pre_persistence

import (
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Codec interface {
	Marshal(proto.Message) ([]byte, types.Error)
	Unmarshal([]byte, proto.Message) types.Error
	ToAny(proto.Message) (*anypb.Any, types.Error)
	FromAny(*anypb.Any) (proto.Message, types.Error)
}

var _ Codec = &ProtoCodec{}

type ProtoCodec struct{}

func (p *ProtoCodec) Marshal(message proto.Message) ([]byte, types.Error) {
	bz, err := proto.Marshal(message)
	if err != nil {
		return nil, types.ErrProtoMarshal(err)
	}
	return bz, nil
}

func (p *ProtoCodec) Unmarshal(b []byte, message proto.Message) types.Error {
	err := proto.Unmarshal(b, message)
	if err != nil {
		return types.ErrProtoUnmarshal(err)
	}
	return nil
}

func (p *ProtoCodec) ToAny(message proto.Message) (*anypb.Any, types.Error) {
	any, err := anypb.New(message)
	if err != nil {
		return nil, types.ErrProtoNewAny(err)
	}
	return any, nil
}

func (p *ProtoCodec) FromAny(any *anypb.Any) (proto.Message, types.Error) {
	msg, err := anypb.UnmarshalNew(any, proto.UnmarshalOptions{})
	if err != nil {
		return nil, types.ErrProtoUnmarshal(err)
	}
	return msg, nil
}

func Cdc() Codec {
	return &ProtoCodec{}
}
