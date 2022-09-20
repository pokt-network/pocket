package modules

import (
	"github.com/pokt-network/pocket/shared/debug"
	"google.golang.org/protobuf/types/known/anypb"
)

type ValidatorMap map[string]Actor // TODO (Drewsky) deprecate Validator map or populate from persistence module

type ConsensusModule interface {
	Module

	// Consensus Engine
	HandleMessage(*anypb.Any) error
	HandleDebugMessage(*debug.DebugMessage) error

	// Consensus State
	CurrentHeight() uint64
	AppHash() string            // DISCUSS: Why not call this a BlockHash or StateHash? Should it be a []byte or string?
	ValidatorMap() ValidatorMap // TODO: This needs to be dynamically updated during various operations and network changes.
}
