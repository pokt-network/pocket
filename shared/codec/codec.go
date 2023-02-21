package codec

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// CONSIDERATION: Use generics in place of `proto.Message` in the interface below so
//
//	every caller does not need to do in place casting.
type Codec interface {
	Marshal(proto.Message) ([]byte, error)
	Unmarshal([]byte, proto.Message) error
	ToAny(proto.Message) (*anypb.Any, error)
	FromAny(*anypb.Any) (proto.Message, error)
	Clone(proto.Message) proto.Message
}

var _ Codec = &protoCodec{}

// IMPROVE: Need to define a type similar to `type ProtoAny anypb.Any` so that we are
//
//	referencing protobuf specific types (e.g. anypb.Any) anywhere in the codebase.
type protoCodec struct{}

func GetCodec() Codec {
	return &protoCodec{}
}

// IMPROVE: If/when we move Pocket's `Error` type into a separate package, we can return `ErrProtoMarshal` here
func (p *protoCodec) Marshal(msg proto.Message) ([]byte, error) {
	return proto.MarshalOptions{Deterministic: true}.Marshal(msg)
}

func (p *protoCodec) Unmarshal(bz []byte, msg proto.Message) error {
	return proto.Unmarshal(bz, msg)
}

func (p *protoCodec) ToAny(msg proto.Message) (*anypb.Any, error) {
	return anypb.New(msg)
}

func (p *protoCodec) FromAny(any *anypb.Any) (proto.Message, error) {
	return anypb.UnmarshalNew(any, proto.UnmarshalOptions{})
}

func (p *protoCodec) Clone(msg proto.Message) proto.Message {
	return proto.Clone(msg)
}
