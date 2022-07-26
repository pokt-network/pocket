package modules

type UtilityModule interface {
	Module
	NewContext(height int64) (UtilityContext, error)
}

// DISCUSS_IN_THIS_COMMIT:
// 1. Explain the relationship between a utility module and the utility context and document it.
type UtilityContext interface {
	// Block operations
	GetProposalTransactions(proposer []byte, maxTransactionBytes int, lastBlockByzantineValidators [][]byte) (transactions [][]byte, err error)
	ApplyProposalTransactions(height int64, proposer []byte, transactions [][]byte, lastBlockByzantineValidators [][]byte) (appHash []byte, err error)
	StoreBlock(blockProtoBytes []byte) error

	// Context operations
	ReleaseContext()
	GetPersistenceContext() PersistenceContext
	CommitPersistenceContext() error

	// Validation operations
	CheckTransaction(tx []byte) error
}

type UnstakingActor interface {
	GetAddress() []byte
	GetStakeAmount() string
	GetOutputAddress() []byte
}
