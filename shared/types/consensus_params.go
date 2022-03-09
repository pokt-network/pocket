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

	// PaceMaker
	PaceMakerParams PaceMakerParams `json:"pace_maker"`
}

type PaceMakerParams struct {
	TimeoutMSec      uint64 `json:"timeout_msec"`
	RetryTimeoutMSec uint64 `json:"retry_timeout_msec"`  // TODO: Not used yet.
	MaxTimeoutMSec   uint64 `json:"max_timeout_msec"`    // TODO: Not used yet.
	MinBlockFreqMSec uint64 `json:"min_block_freq_msec"` // TODO: Not used yet.
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

	if err := c.PaceMakerParams.Validate(); err != nil {
		return fmt.Errorf("Pacemaker params invalid: %w", err)
	}

	return nil
}

func (p *PaceMakerParams) Validate() error {
	if p.TimeoutMSec <= 0 {
		return fmt.Errorf("TimeoutMSec must be a positive integer")
	}

	if p.RetryTimeoutMSec <= 0 {
		return fmt.Errorf("RetryTimeoutMSec must be a positive integer")
	}

	if p.MaxTimeoutMSec <= 0 {
		return fmt.Errorf("MaxTimeoutMSec must be a positive integer")
	}

	if p.MinBlockFreqMSec <= 0 {
		return fmt.Errorf("MinBlockFreqMSec must be a positive integer")
	}

	return nil
}
