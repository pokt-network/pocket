package ibc

import (
	"fmt"

	"github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

func CreateUpdateStoreMessage(prefix coreTypes.CommitmentPrefix, key, value []byte) *types.IbcMessage {
	return &types.IbcMessage{
		Event: &types.IbcMessage_Update{
			Update: &types.UpdateIbcStore{
				Prefix: prefix,
				Key:    key,
				Value:  value,
			},
		},
	}
}

func CreatePruneStoreMessage(prefix coreTypes.CommitmentPrefix, key []byte) *types.IbcMessage {
	return &types.IbcMessage{
		Event: &types.IbcMessage_Prune{
			Prune: &types.PruneIbcStore{
				Prefix: prefix,
				Key:    key,
			},
		},
	}
}

func ConvertIBCMessageToTx(ibcMessage *types.IbcMessage) (*coreTypes.Transaction, error) {
	var anyMsg *anypb.Any
	var err error
	switch event := ibcMessage.Event.(type) {
	case *types.IbcMessage_Update:
		anyMsg, err = codec.GetCodec().ToAny(event.Update)
	case *types.IbcMessage_Prune:
		anyMsg, err = codec.GetCodec().ToAny(event.Prune)
	default:
		return nil, coreTypes.ErrUnknownIBCMessageType(fmt.Sprintf("%T", event))
	}
	if err != nil {
		return nil, err
	}
	return &coreTypes.Transaction{
		Msg:   anyMsg,
		Nonce: fmt.Sprintf("%d", crypto.GetNonce()),
	}, nil
}
