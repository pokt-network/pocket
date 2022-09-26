package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/consensus_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	"github.com/pokt-network/pocket/shared/debug"
	"google.golang.org/protobuf/types/known/anypb"
)

// TODO(olshansky): deprecate ValidatorMap or populate from persistence module
type ValidatorMap map[string]Actor

type ConsensusModule interface {
	Module

	// Consensus Engine Handlers
	HandleMessage(*anypb.Any) error
	HandleDebugMessage(*debug.DebugMessage) error

	// Consensus State Accessors
	CurrentHeight() uint64
	AppHash() string            // DISCUSS: Why not call this a BlockHash or StateHash? Should it be a []byte or string?
	ValidatorMap() ValidatorMap // TODO: This needs to be dynamically updated during various operations and network changes.
}
