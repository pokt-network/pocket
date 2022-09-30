package codec

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Codec interface {
	Marshal(proto.Message) ([]byte, error)
	Unmarshal([]byte, proto.Message) error
	ToAny(proto.Message) (*anypb.Any, error)
	FromAny(*anypb.Any) (proto.Message, error)
}

var _ Codec = &ProtoCodec{}

type ProtoCodec struct{}

func (p *ProtoCodec) Marshal(message proto.Message) ([]byte, error) {
	bz, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func (p *ProtoCodec) Unmarshal(bz []byte, message proto.Message) error {
	err := proto.Unmarshal(bz, message)
	if err != nil {
		return err
	}
	return nil
}

func (p *ProtoCodec) ToAny(message proto.Message) (*anypb.Any, error) {
	any, err := anypb.New(message)
	if err != nil {
		return nil, err
	}
	return any, nil
}

func (p *ProtoCodec) FromAny(any *anypb.Any) (proto.Message, error) {
	msg, err := anypb.UnmarshalNew(any, proto.UnmarshalOptions{})
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// DISCUSS: Retrieve this from the utility module via the application specific bus?
// There are some parts of the code that does not have access to the bus;
// Example: txIndexer
func GetCodec() Codec {
	return &ProtoCodec{}
}
