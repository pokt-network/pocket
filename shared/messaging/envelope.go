package messaging

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

// PackMessage returns a *PocketEnvelope after having packed the message supplied as an argument
func PackMessage(message proto.Message) (*PocketEnvelope, error) {
	anyMsg, err := anypb.New(message)
	if err != nil {
		return nil, err
	}
	return &PocketEnvelope{Content: anyMsg}, nil
}

// UnpackMessage extracts the message inside the PocketEnvelope decorating it with typing information
func UnpackMessage(envelope *PocketEnvelope) (protoreflect.ProtoMessage, error) {
	anyMsg := envelope.Content
	msg, err := anypb.UnmarshalNew(anyMsg, proto.UnmarshalOptions{})
	if err != nil {
		return nil, err
	}
	return msg, nil
}
