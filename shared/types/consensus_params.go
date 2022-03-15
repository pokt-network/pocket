package types

import "fmt"

type ConsensusParams struct {
	// Mempool
	MaxMempoolBytes uint64 `json:"max_mempool_bytes"`

	// Block
	MaxBlockBytes       uint64 `json:"max_block_bytes"`
	MaxTransactionBytes uint64 `json:"max_transaction_bytes"`
}

func (c *ConsensusParams) Validate() error {
	if c.MaxMempoolBytes <= 0 {
		return fmt.Errorf("MaxMempoolBytes must be a positive integer")
	}

	if c.MaxBlockBytes <= 0 {
		return fmt.Errorf("MaxBlockBytes must be a positive integer")
	}

	if c.MaxTransactionBytes <= 0 {
		return fmt.Errorf("MaxTransactionBytes must be a positive integer")
	}

	return nil
}
