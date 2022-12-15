package modules

import (
	"google.golang.org/protobuf/types/known/anypb"
)

//go:generate mockgen -source=$GOFILE -destination=./mocks/utility_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

type UtilityModule interface {
	Module
	ConfigurableModule

	HandleMessage(*anypb.Any) error

	// Creates a utilityContext with an underlying read-write persistenceContext; only 1 can exist at a time
	NewContext(height int64) (UtilityContext, error)

	// Basic Transaction validation. SIDE EFFECT: Adds the transaction to the mempool if valid.
	CheckTransaction(tx []byte) error
}

// Interface defining the context within which the node can operate with the utility layer.
// Operations in the context of a UtilityContext are isolated from other operations and
// other utility contexts until committed and released, enabling parallelizability along other
// operations.
type UtilityContext interface {
	// Block operations

	// This function is intended to be called by any type of node during state transitions. For example,
	// both block proposers and verifiers will use it to create a context (before finalizing it) during consensus,
	// and all verifiers will call it during state sync.
	SetProposalBlock(blockHash string, proposerAddr []byte, transactions [][]byte) error
	// Reaps the mempool for transactions to be proposed in a new block, and applies them to this
	// context; intended to be used by the block proposer.
	CreateAndApplyProposalBlock(proposer []byte, maxTransactionBytes int) (stateHash string, transactions [][]byte, err error)
	// Applies the proposed local state (i.e. the transactions in the current context); intended to be used by
	// the block verifiers (i.e. non proposers)..
	ApplyBlock() (stateHash string, err error)

	// TECHDEBT: `CreateAndApplyProposalBlock` and `ApplyBlock` should be be refactored into a
	// `GetProposalBlock` and `ApplyProposalBlock` functions

	// Context operations

	// Releases the utility context and any underlying contexts it references
	Release() error
	// State commitment of the current context
	Commit(quorumCert []byte) error
	// Returns the read-write persistence context initialized by this utility context
	GetPersistenceContext() PersistenceRWContext
}

type UnstakingActor interface {
	GetAddress() []byte
	GetStakeAmount() string
	GetOutputAddress() []byte
}
