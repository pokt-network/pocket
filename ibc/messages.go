package ibc

import (
	"github.com/pokt-network/pocket/ibc/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

func CreateUpdateStoreMessage(prefix coreTypes.CommitmentPrefix, key, value []byte) *types.IBCMessage {
	return &types.IBCMessage{
		Msg: &types.IBCMessage_Update{
			Update: &types.UpdateIBCStore{
				Prefix: prefix,
				Key:    key,
				Value:  value,
			},
		},
	}
}

func CreatePruneStoreMessage(prefix coreTypes.CommitmentPrefix, key []byte) *types.IBCMessage {
	return &types.IBCMessage{
		Msg: &types.IBCMessage_Prune{
			Prune: &types.PruneIBCStore{
				Prefix: prefix,
				Key:    key,
			},
		},
	}
}
