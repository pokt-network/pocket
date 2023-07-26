package types

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

func CreateUpdateStoreMessage(key, value []byte) *IBCMessage {
	return &IBCMessage{
		Event: &IBCMessage_Update{
			Update: &UpdateIBCStore{
				Key:   key,
				Value: value,
			},
		},
	}
}

func CreatePruneStoreMessage(key []byte) *IBCMessage {
	return &IBCMessage{
		Event: &IBCMessage_Prune{
			Prune: &PruneIBCStore{
				Key: key,
			},
		},
	}
}

func ConvertIBCMessageToTx(ibcMessage *IBCMessage) (*coreTypes.Transaction, error) {
	var anyMsg *anypb.Any
	var err error
	switch event := ibcMessage.Event.(type) {
	case *IBCMessage_Update:
		anyMsg, err = codec.GetCodec().ToAny(event.Update)
	case *IBCMessage_Prune:
		anyMsg, err = codec.GetCodec().ToAny(event.Prune)
	default:
		return nil, coreTypes.ErrIBCUnknownMessageType(fmt.Sprintf("%T", event))
	}
	if err != nil {
		return nil, err
	}
	return &coreTypes.Transaction{
		Msg:   anyMsg,
		Nonce: fmt.Sprintf("%d", crypto.GetNonce()),
	}, nil
}
