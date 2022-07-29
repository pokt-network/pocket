package modules

import (
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"google.golang.org/protobuf/types/known/anypb"
)

type ValidatorMap map[string]*typesGenesis.Validator

type ConsensusModule interface {
	Module

	// Consensus Engine
	HandleMessage(*anypb.Any) error
	HandleDebugMessage(*types.DebugMessage) error

	// Consensus State
	// DISCUSS(team): This is a temporary solutions to retrieve heights until we can do so from the pre-persistence module
	// At the moment, the pre-persistence module has a `GetLatestBlockHeight` method in its context (`PrePersistenceContext`), but context methods
	// are not exposed to other modules through the bus, thus is unusable outside persistence so far.
	// TODO(team): Remember to remove this method from the consensus module once this functionality is offered to other modules through persistence
	BlockHeight() uint64
	AppHash() string            // DISCUSS: Why not call this a BlockHash or StateHash? Should it be a []byte or string?
	ValidatorMap() ValidatorMap // TODO: This needs to be dynamically updated during various operations and network changes.
}
