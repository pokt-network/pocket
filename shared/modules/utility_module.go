package modules

type UtilityModule interface {
	Module
	NewContext(height int64) (UtilityContext, error)
}

// DOCUMENT_IN_THIS_COMMIT: Explain the relationship between a utility module and the utility context
type UtilityContext interface {
	// Block operations
	GetProposalTransactions(proposer []byte, maxTransactionBytes int, lastBlockByzantineValidators [][]byte) (transactions [][]byte, err error)
	ApplyProposalTransactions(height int64, proposer []byte, transactions [][]byte, lastBlockByzantineValidators [][]byte) (appHash []byte, err error)

	// Context operations
	ReleaseContext()
	GetPersistenceContext() PersistenceContext

	// Validation operations
	CheckTransaction(tx []byte) error
}

type UnstakingActor interface {
	GetAddress() []byte
	GetStakeAmount() string
	GetOutputAddress() []byte
}
