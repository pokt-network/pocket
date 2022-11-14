package messaging

import (
	"log"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

func PackMessage(message proto.Message) (*PocketEnvelope, error) {
	anyMsg, err := anypb.New(message)
	if err != nil {
		return nil, err
	}
	return &PocketEnvelope{Content: anyMsg}, nil
}

func MustPackMessage(message proto.Message) *PocketEnvelope {
	anyMsg, err := anypb.New(message)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create Any proto: %v", err)
	}
	return &PocketEnvelope{Content: anyMsg}
}

func UnpackMessage(envelope *PocketEnvelope) (protoreflect.ProtoMessage, error) {
	anyMsg := envelope.Content
	msg, err := anypb.UnmarshalNew(anyMsg, proto.UnmarshalOptions{})
	if err != nil {
		return nil, err
	}
	return msg, nil
}
