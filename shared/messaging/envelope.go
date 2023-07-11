package messaging

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// PackMessage returns a *PocketEnvelope after having packed the message supplied as an argument
func PackMessage(message proto.Message) (*PocketEnvelope, error) {
	anyMsg, err := anypb.New(message)
	if err != nil {
		return nil, err
	}
	return &PocketEnvelope{
		Content: anyMsg,
		Nonce:   cryptoPocket.GetNonce(),
	}, nil
}

// UnpackMessage extracts the message inside the PocketEnvelope decorating it with typing information
func UnpackMessage[T protoreflect.ProtoMessage](envelope *PocketEnvelope) (T, error) {
	anyMsg := envelope.Content
	msg, err := anypb.UnmarshalNew(anyMsg, proto.UnmarshalOptions{})
	if err != nil {
		return any(nil).(T), err
	}
	return any(msg).(T), nil
}
