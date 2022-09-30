package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/consensus_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	"github.com/pokt-network/pocket/shared/debug"
	"google.golang.org/protobuf/types/known/anypb"
)

type ValidatorMap map[string]Actor // TODO : Validator map should be a shared interface not a structure

type ConsensusModule interface {
	Module

	// Consensus Engine
	HandleMessage(*anypb.Any) error
	HandleDebugMessage(*debug.DebugMessage) error

	// Consensus State
	CurrentHeight() uint64
	AppHash() string // DISCUSS: Why not call this a BlockHash or StateHash? Should it be a []byte or string?
	ValidatorMap() ValidatorMap
}
