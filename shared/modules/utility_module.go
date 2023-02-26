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

	// Creates a `utilityContext` with an underlying read-write `persistenceContext`; only 1 of which can exist at a time
	NewContext(height int64) (UtilityContext, error)

	// Basic Transaction validation & addition to the utility's module mempool upon successful validation
	HandleTransaction(tx []byte) error

	// Returns the utility module's mempool of transactions gossiped throughout the network
	GetMempool() mempool.TXMempool

	// General purpose handler of utility specific messages used for utility specific business logic
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
	// Applies the proposed local state (i.e. the transactions in the current context).
	// Only intended to be used by the block verifiers (i.e. replicas).
	ApplyBlock() (stateHash string, err error)

	// Context operations

	// Releases the utility context and any underlying contexts it references
	Release() error
	// Commit the current utility context (along with its underlying persistence context) to disk
	Commit(quorumCert []byte) error
}

// TECHDEBT: Remove this interface from `shared/modules`
type UnstakingActor interface {
	GetAddress() []byte
	GetOutputAddress() []byte
}
