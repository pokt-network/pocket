//go:build test

package modules

import (
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
)

// ConsensusDebugModule exposes functionality used for testing & development purposes.
// Not to be used in production.
type ConsensusDebugModule interface {
	HandleDebugMessage(*messaging.DebugMessage) error

	SetHeight(uint64)
	SetRound(uint64)
	SetStep(uint8) // REFACTOR: This should accept typesCons.HotstuffStep
	SetBlock(*types.Block)

	SetUtilityUnitOfWork(UtilityUnitOfWork)

	// REFACTOR: This should accept typesCons.HotstuffStep and return typesCons.NodeId.
	GetLeaderForView(height, round uint64, step uint8) (leaderId uint64)
}
