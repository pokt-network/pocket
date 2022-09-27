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
	GetProposalTransactions(proposer []byte, maxTransactionBytes int, lastBlockByzantineValidators [][]byte) (transactions [][]byte, err error)
	ApplyBlock(height int64, proposer []byte, transactions [][]byte, lastBlockByzantineValidators [][]byte) (appHash []byte, err error)
	StoreBlock(blockProtoBytes []byte) error

	// Context operations
	ReleaseContext()
	GetPersistenceContext() PersistenceRWContext
	CommitPersistenceContext() error

	// Validation operations
	CheckTransaction(tx []byte) error
}

type UnstakingActor interface {
	GetAddress() []byte
	GetStakeAmount() string
	GetOutputAddress() []byte
}
