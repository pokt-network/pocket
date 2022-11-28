package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/utility_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

type UtilityModule interface {
	Module
	ConfigurableModule

	NewContext(height int64) (UtilityContext, error)
}

// Interface defining the context within which the node can operate with the utility layer.
// Operations in the context of a UtilityContext are isolated from other operations and
// other utility contexts until committed and released, enabling parallelizability along other
// operations.
type UtilityContext interface {
	// Block operations

	// Reaps the mempool for transactions to be proposed in a new block, and applies them to this
	// context; intended to be used by the block proposer.
	CreateAndApplyProposalBlock(proposer []byte, maxTransactionBytes int) (appHash []byte, transactions [][]byte, err error)
	// Applies the transactions in the local state to the current context; intended to be used by
	// the block verifiers (i.e. non proposers)..
	ApplyBlock() (appHash []byte, err error)

	// Context operations
	Release() error                 // Releases the utility context and any underlying contexts it references
	Commit(quorumCert []byte) error // State commitment of the current context
	GetPersistenceContext() PersistenceRWContext

	// Validation operations
	CheckTransaction(tx []byte) error
}

type UnstakingActor interface {
	GetAddress() []byte
	GetStakeAmount() string
	GetOutputAddress() []byte
}
