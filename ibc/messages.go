package ibc

import (
	"github.com/pokt-network/pocket/ibc/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
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
