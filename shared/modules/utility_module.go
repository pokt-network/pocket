package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/utility_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	"github.com/pokt-network/pocket/shared/mempool"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	UtilityModuleName = "utility"
)

type UtilityModule interface {
	Module

	// NewContext creates a `utilityContext` with an underlying read-write `persistenceContext` (only 1 of which can exist at a time)
	NewContext(height int64) (UtilityContext, error)

	// HandleTransaction does basic `Transaction` validation & adds it to the utility's module mempool if valid
	HandleTransaction(tx []byte) error

	// GetMempool returns the utility module's mempool of transactions gossiped throughout the network
	GetMempool() mempool.TXMempool

	// HandleUtilityMessage is a general purpose handler of utility-specific messages used for utility-specific business logic.
	// It is useful for handling messages from the utility module's of other nodes that do not directly affect the state.
	// IMPROVE: Find opportunities to break this apart as the module matures.
	HandleUtilityMessage(*anypb.Any) error
}

// TECHDEBT: `CreateAndApplyProposalBlock` and `ApplyBlock` should be be refactored into a
//           `GetProposalBlock` and `ApplyProposalBlock` functions

// The context within which the node can operate with the utility layer.
type UtilityContext interface {
	// Block operations

	// This function is intended to be called by any type of node during state transitions.
	// For example, both block proposers and replicas/verifiers will use it to create a
	// context (before finalizing it) during consensus, and all verifiers will call it during state sync.
	// TODO: Replace []byte with semantic type
	SetProposalBlock(blockHash string, proposerAddr []byte, txs [][]byte) error

	// Reaps the mempool for transactions to be proposed in a new block, and applies them to this
	// context.
	// Only intended to be used by the block proposer.
	// TODO: Replace []byte with semantic type
	CreateAndApplyProposalBlock(proposer []byte, maxTxBytes int) (stateHash string, txs [][]byte, err error)

	// ApplyBlock applies the context's in-memory proposed state (i.e. the txs in this context).
	// Only intended to be used by the block verifiers (i.e. replicas).
	ApplyBlock() (stateHash string, err error)

	// Context operations

	// Release releases this utility context and any underlying contexts it references
	Release() error

	// Commit commits this utility context along with any underlying contexts (e.g. persistenceContext) it references
	Commit(quorumCert []byte) error
}

// TECHDEBT: Remove this interface from `shared/modules` and use the `Actor` protobuf type instead
// There will need to be some documentation or indicator that the Actor struct returned may not be
// fully hydrated. Alternatively, we could eat the performance cost and just hydrate the entire struct
// which may be simpler and clearer.
type UnstakingActor interface {
	GetAddress() []byte
	GetOutputAddress() []byte
}
