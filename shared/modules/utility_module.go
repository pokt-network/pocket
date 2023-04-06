package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/utility_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	"github.com/pokt-network/pocket/shared/mempool"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	UtilityModuleName = "utility"
)

// TECHDEBT: Replace []byte with semantic type (addresses, transactions, etc...)

type UtilityModule interface {
	Module

	// NewUnitOfWork creates a `utilityUnitOfWork` used to allow atomicity and commit/rollback functionality
	NewUnitOfWork(height int64) (UtilityUnitOfWork, error)

	// HandleTransaction does basic `Transaction` validation & adds it to the utility's module mempool if valid
	HandleTransaction(tx []byte) error

	// GetMempool returns the utility module's mempool of transactions gossiped throughout the network
	GetMempool() mempool.TXMempool

	// HandleUtilityMessage is a general purpose handler of utility-specific messages used for utility-specific business logic.
	// It is useful for handling messages from the utility module's of other nodes that do not directly affect the state.
	// IMPROVE: Find opportunities to break this apart as the module matures.
	HandleUtilityMessage(*anypb.Any) error
}

// TECHDEBT: Remove this interface from `shared/modules` and use the `Actor` protobuf type instead
// There will need to be some documentation or indicator that the Actor struct returned may not be
// fully hydrated. Alternatively, we could eat the performance cost and just hydrate the entire struct
// which may be simpler and clearer.
type UnstakingActor interface {
	GetAddress() []byte
	GetOutputAddress() []byte
}

// CONSIDERATION: Consider removing `Utility` from `UtilityUnitOfWork` altogether

// TECHDEBT(@deblasis): `CreateAndApplyProposalBlock` and `ApplyBlock` should be be refactored into a
// `GetProposalBlock` and `ApplyProposalBlock` functions

// UtilityUnitOfWork is a unit of work (https://martinfowler.com/eaaCatalog/unitOfWork.html) that allows for atomicity and commit/rollback functionality
type UtilityUnitOfWork interface {
	IntegratableModule

	// SetProposalBlock updates the utility unit of work with the proposed state transition.
	// It does not apply, validate or commit the changes.
	// For example, it can be use during state sync to set a proposed state transition before validation.
	// TODO: Investigate a way to potentially simplify the interface by removing this function.
	// TODO: @deblasis: there's still some mix and match between blockHash and stateHash
	SetProposalBlock(blockHash string, proposerAddr []byte, txs [][]byte) error

	// ApplyBlock applies the context's in-memory proposed state (i.e. the txs in this context).
	// Only intended to be used by the block verifiers (i.e. replicas).
	// NOTE: this is called by the replica OR by the leader when `prepareQc` is not `nil`
	ApplyBlock() (stateHash string, txs [][]byte, err error)

	// Release releases this utility unit of work and any underlying contexts it references
	Release() error

	// Commit commits this utility unit of work along with any underlying contexts (e.g. persistenceContext) it references
	Commit(quorumCert []byte) error
}

type LeaderUtilityUnitOfWork interface {
	UtilityUnitOfWork

	// CreateAndApplyProposalBlock reaps the mempool for txs to be proposed in a new block, and
	// applies them to this context after validation.
	// TODO: #508 new signature
	CreateAndApplyProposalBlock(proposer []byte, maxTxBytes uint64) (stateHash string, txs [][]byte, err error)
}

type ReplicaUtilityUnitOfWork interface {
	UtilityUnitOfWork
}
