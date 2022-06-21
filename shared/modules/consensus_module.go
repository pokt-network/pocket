package modules

import (
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/types/known/anypb"
)

type ConsensusModule interface {
	Module
	HandleMessage(*anypb.Any) error
	HandleDebugMessage(*types.DebugMessage) error
}
