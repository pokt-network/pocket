package types

import "fmt"

type ConsensusParams struct {
	// Mempool
	MaxMempoolBytes uint64 `json:"max_mempool_bytes"`

	// Block
	MaxBlockBytes       uint64 `json:"max_block_bytes"`
	MaxTransactionBytes uint64 `json:"max_transaction_bytes"`

	// VRF
	VRFKeyRefreshFreqBlock uint64 `json:"vrf_key_refresh_freq_block"`
	VRFKeyValidityBlock    uint64 `json:"vrf_key_validity_block"`
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

	if c.VRFKeyRefreshFreqBlock <= 0 {
		return fmt.Errorf("VRFKeyRefreshFreqBlock must be a positive integer")
	}

	if c.VRFKeyValidityBlock <= 0 {
		return fmt.Errorf("VRFKeyValidityBlock must be a positive integer")
	}

	return nil
}
