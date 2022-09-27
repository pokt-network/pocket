package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/utility_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

type UtilityModule interface {
	Module

	NewContext(height int64) (UtilityContext, error)
}

// Interface defining the context within which the node can operate with the utility layer.
// Operations in the context of a UtilityContext are isolated from other operations and
// other utility contexts until committed and released, enabling parallelizability along other
// operations.
type UtilityContext interface {
	// Block operations

	// Reaps the mempool for transactions that are ready to be proposed in a new block
	GetProposalTransactions(proposer []byte, maxTransactionBytes int, lastBlockByzantineValidators [][]byte) (transactions [][]byte, err error)
	// Applies the transactions to an ephemeral state in the utility & underlying persistence context;similar to `SafeNode` in the Hotstuff whitepaper.
	ApplyBlock(height int64, proposer []byte, transactions [][]byte, lastBlockByzantineValidators [][]byte) (appHash []byte, err error)

	// Context operations
	ReleaseContext() error
	CommitContext(quorumCert []byte) error

	// Validation operations
	CheckTransaction(tx []byte) error
}

type UnstakingActor interface {
	GetAddress() []byte
	GetStakeAmount() string
	GetOutputAddress() []byte
}
