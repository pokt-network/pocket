package modules

type UnstakingActor interface {
	GetAddress() []byte
	GetStakeAmount() string
	GetOutputAddress() []byte
}

type UtilityContextInterface interface {
	ReleaseContext()
	GetPersistanceContext() PersistenceContext
	CheckTransaction(tx []byte) error
	GetTransactionsForProposal(proposer []byte, maxTransactionBytes int, lastBlockByzantineValidators [][]byte) (transactions [][]byte, err error)
	ApplyBlock(Height int64, proposer []byte, transactions [][]byte, lastBlockByzantineValidators [][]byte) (appHash []byte, err error)
}

type UtilityModule interface {
	PocketModule

	NewUtilityContextWrapper(height int64) (UtilityContextInterface, error) // INTEGRATION_TEMP: need to move `types.Errors` to shared

	// // Message Handling
	// HandleTransaction(*context.PocketContext, *typespb.Transaction) error
	// HandleEvidence(*context.PocketContext, *typespb.Evidence) error

	// // Block Application
	// ReapMempool(*context.PocketContext) ([]*typespb.Transaction, error)
	// BeginBlock(*context.PocketContext) error
	// DeliverTx(*context.PocketContext, *typespb.Transaction) error
	// EndBlock(*context.PocketContext) error
}
