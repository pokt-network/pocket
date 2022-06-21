package types

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Codec interface { // TODO (Team) move to shared. Possibly rename
	Marshal(proto.Message) ([]byte, Error)
	Unmarshal([]byte, proto.Message) Error
	ToAny(proto.Message) (*anypb.Any, Error)
	FromAny(*anypb.Any) (proto.Message, Error)
}

var _ Codec = &ProtoCodec{}

type ProtoCodec struct{}

func (p *ProtoCodec) Marshal(message proto.Message) ([]byte, Error) {
	bz, err := proto.Marshal(message)
	if err != nil {
		return nil, ErrProtoMarshal(err)
	}
	return bz, nil
}

func (p *ProtoCodec) Unmarshal(bz []byte, message proto.Message) Error {
	err := proto.Unmarshal(bz, message)
	if err != nil {
		return ErrProtoUnmarshal(err)
	}
	return nil
}

func (p *ProtoCodec) ToAny(message proto.Message) (*anypb.Any, Error) {
	any, err := anypb.New(message)
	if err != nil {
		return nil, ErrProtoNewAny(err)
	}
	return any, nil
}

func (p *ProtoCodec) FromAny(any *anypb.Any) (proto.Message, Error) {
	msg, err := anypb.UnmarshalNew(any, proto.UnmarshalOptions{})
	if err != nil {
		return nil, ErrProtoUnmarshal(err)
	}
	return msg, nil
}

func GetCodec() Codec {
	return &ProtoCodec{}
}
