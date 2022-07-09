package modules

import (
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/types/known/anypb"
)

type ConsensusModule interface {
	Module
	HandleMessage(*anypb.Any) error
	HandleDebugMessage(*types.DebugMessage) error

	// DISCUSS(team): This is a temporary solutions to retrieve heights until we can do so from the pre-persistence module
	// At the moment, the pre-persistence module has a `GetLatestBlockHeight` method in its context (`PrePersistenceContext`), but context methods
	// are not exposed to other modules through the bus, thus is unusable outside persistence so far.
	// TODO(team): Remember to remove this method from the consensus module once this functionality is offered to other modules through persistence
	GetBlockHeight() uint64
}
