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
	BlockHeight() uint64
	AppHash() string            // DISCUSS: Why not call this a BlockHash or StateHash? Should it be a []byte or string?
	ValidatorMap() ValidatorMap // TODO: Need to update this on every validator pause/stake/unstake/etc.
	// TotalVotingPower() uint64                         // TODO: Need to update this on every send transaction.
}
