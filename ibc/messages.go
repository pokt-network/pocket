package ibc

import (
	"fmt"

	"github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

func CreateUpdateStoreMessage(key, value []byte) *types.IBCMessage {
	return &types.IBCMessage{
		Event: &types.IBCMessage_Update{
			Update: &types.UpdateIBCStore{
				Key:   key,
				Value: value,
			},
		},
	}
}

func CreatePruneStoreMessage(key []byte) *types.IBCMessage {
	return &types.IBCMessage{
		Event: &types.IBCMessage_Prune{
			Prune: &types.PruneIBCStore{
				Key: key,
			},
		},
	}
}

func ConvertIBCMessageToTx(ibcMessage *types.IBCMessage) (*coreTypes.Transaction, error) {
	var anyMsg *anypb.Any
	var err error
	switch event := ibcMessage.Event.(type) {
	case *types.IBCMessage_Update:
		anyMsg, err = codec.GetCodec().ToAny(event.Update)
	case *types.IBCMessage_Prune:
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
