package test_artifacts

import (
	"math/big"

	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/utils"
)

// TECHDEBT: This entire file should be re-scoped.
// The test suite should be customizable but the default params are a good starting point.

var (
	DefaultChains              = []string{"0001"}
	DefaultServiceURL          = ""
	DefaultStakeAmount         = big.NewInt(1000000000000)
	DefaultStakeAmountString   = utils.BigIntToString(DefaultStakeAmount)
	DefaultAccountAmount       = big.NewInt(100000000000000)
	DefaultAccountAmountString = utils.BigIntToString(DefaultAccountAmount)
	DefaultPauseHeight         = int64(-1) // pauseHeight=-1 implies not paused
	DefaultUnstakingHeight     = int64(-1) // unstakingHeight=-1 implies not unstaking
	DefaultChainID             = "testnet"
	ServiceURLFormat           = "node%d.consensus:42069"
	DefaultMaxBlockBytes       = uint64(4000000)
	DefaultParamsOwner, _      = crypto.NewPrivateKey("ff538589deb7f28bbce1ba68b37d2efc0eaa03204b36513cf88422a875559e38d6cbe0430ddd85a5e48e0c99ef3dea47bf0d1a83c6e6ad1640f72201dc8a0120")
)

func DefaultParams() *genesis.Params {
	return &genesis.Params{
		BlocksPerSession:                     1,
		AppMinimumStake:                      utils.BigIntToString(big.NewInt(15000000000)),
		AppMaxChains:                         15,
		AppSessionTokensMultiplier:           100,
		AppUnstakingBlocks:                   2016,
		AppMinimumPauseBlocks:                4,
		AppMaxPauseBlocks:                    672,
		ServicerMinimumStake:                 utils.BigIntToString(big.NewInt(15000000000)),
		ServicerMaxChains:                    15,
		ServicerUnstakingBlocks:              2016,
		ServicerMinimumPauseBlocks:           4,
		ServicerMaxPauseBlocks:               672,
		ServicersPerSession:                  24,
		WatcherMinimumStake:                  utils.BigIntToString(big.NewInt(15000000000)),
		WatcherMaxChains:                     15,
		WatcherUnstakingBlocks:               2016,
		WatcherMinimumPauseBlocks:            4,
		WatcherMaxPauseBlocks:                672,
		WatcherPerSession:                    1,
		ValidatorMinimumStake:                utils.BigIntToString(big.NewInt(15000000000)),
		ValidatorUnstakingBlocks:             2016,
		ValidatorMinimumPauseBlocks:          4,
		ValidatorMaxPauseBlocks:              672,
		ValidatorMaximumMissedBlocks:         5,
		ValidatorMaxEvidenceAgeInBlocks:      8,
		ProposerPercentageOfFees:             10,
		MissedBlocksBurnPercentage:           1,
		DoubleSignBurnPercentage:             5,
		MessageDoubleSignFee:                 utils.BigIntToString(big.NewInt(10000)),
		MessageSendFee:                       utils.BigIntToString(big.NewInt(10000)),
		MessageStakeWatcherFee:               utils.BigIntToString(big.NewInt(10000)),
		MessageEditStakeWatcherFee:           utils.BigIntToString(big.NewInt(10000)),
		MessageUnstakeWatcherFee:             utils.BigIntToString(big.NewInt(10000)),
		MessagePauseWatcherFee:               utils.BigIntToString(big.NewInt(10000)),
		MessageUnpauseWatcherFee:             utils.BigIntToString(big.NewInt(10000)),
		MessageWatcherPauseServicerFee:       utils.BigIntToString(big.NewInt(10000)),
		MessageTestScoreFee:                  utils.BigIntToString(big.NewInt(10000)),
		MessageProveTestScoreFee:             utils.BigIntToString(big.NewInt(10000)),
		MessageStakeAppFee:                   utils.BigIntToString(big.NewInt(10000)),
		MessageEditStakeAppFee:               utils.BigIntToString(big.NewInt(10000)),
		MessageUnstakeAppFee:                 utils.BigIntToString(big.NewInt(10000)),
		MessagePauseAppFee:                   utils.BigIntToString(big.NewInt(10000)),
		MessageUnpauseAppFee:                 utils.BigIntToString(big.NewInt(10000)),
		MessageStakeValidatorFee:             utils.BigIntToString(big.NewInt(10000)),
		MessageEditStakeValidatorFee:         utils.BigIntToString(big.NewInt(10000)),
		MessageUnstakeValidatorFee:           utils.BigIntToString(big.NewInt(10000)),
		MessagePauseValidatorFee:             utils.BigIntToString(big.NewInt(10000)),
		MessageUnpauseValidatorFee:           utils.BigIntToString(big.NewInt(10000)),
		MessageStakeServicerFee:              utils.BigIntToString(big.NewInt(10000)),
		MessageEditStakeServicerFee:          utils.BigIntToString(big.NewInt(10000)),
		MessageUnstakeServicerFee:            utils.BigIntToString(big.NewInt(10000)),
		MessagePauseServicerFee:              utils.BigIntToString(big.NewInt(10000)),
		MessageUnpauseServicerFee:            utils.BigIntToString(big.NewInt(10000)),
		MessageChangeParameterFee:            utils.BigIntToString(big.NewInt(10000)),
		AclOwner:                             DefaultParamsOwner.Address().String(),
		BlocksPerSessionOwner:                DefaultParamsOwner.Address().String(),
		AppMinimumStakeOwner:                 DefaultParamsOwner.Address().String(),
		AppMaxChainsOwner:                    DefaultParamsOwner.Address().String(),
		AppSessionTokensMultiplierOwner:      DefaultParamsOwner.Address().String(),
		AppUnstakingBlocksOwner:              DefaultParamsOwner.Address().String(),
		AppMinimumPauseBlocksOwner:           DefaultParamsOwner.Address().String(),
		AppMaxPausedBlocksOwner:              DefaultParamsOwner.Address().String(),
		ServicerMinimumStakeOwner:            DefaultParamsOwner.Address().String(),
		ServicerMaxChainsOwner:               DefaultParamsOwner.Address().String(),
		ServicerUnstakingBlocksOwner:         DefaultParamsOwner.Address().String(),
		ServicerMinimumPauseBlocksOwner:      DefaultParamsOwner.Address().String(),
		ServicerMaxPausedBlocksOwner:         DefaultParamsOwner.Address().String(),
		ServicersPerSessionOwner:             DefaultParamsOwner.Address().String(),
		WatcherMinimumStakeOwner:             DefaultParamsOwner.Address().String(),
		WatcherMaxChainsOwner:                DefaultParamsOwner.Address().String(),
		WatcherUnstakingBlocksOwner:          DefaultParamsOwner.Address().String(),
		WatcherMinimumPauseBlocksOwner:       DefaultParamsOwner.Address().String(),
		WatcherMaxPausedBlocksOwner:          DefaultParamsOwner.Address().String(),
		WatcherPerSessionOwner:               DefaultParamsOwner.Address().String(),
		ValidatorMinimumStakeOwner:           DefaultParamsOwner.Address().String(),
		ValidatorUnstakingBlocksOwner:        DefaultParamsOwner.Address().String(),
		ValidatorMinimumPauseBlocksOwner:     DefaultParamsOwner.Address().String(),
		ValidatorMaxPausedBlocksOwner:        DefaultParamsOwner.Address().String(),
		ValidatorMaximumMissedBlocksOwner:    DefaultParamsOwner.Address().String(),
		ValidatorMaxEvidenceAgeInBlocksOwner: DefaultParamsOwner.Address().String(),
		ProposerPercentageOfFeesOwner:        DefaultParamsOwner.Address().String(),
		MissedBlocksBurnPercentageOwner:      DefaultParamsOwner.Address().String(),
		DoubleSignBurnPercentageOwner:        DefaultParamsOwner.Address().String(),
		MessageDoubleSignFeeOwner:            DefaultParamsOwner.Address().String(),
		MessageSendFeeOwner:                  DefaultParamsOwner.Address().String(),
		MessageStakeWatcherFeeOwner:          DefaultParamsOwner.Address().String(),
		MessageEditStakeWatcherFeeOwner:      DefaultParamsOwner.Address().String(),
		MessageUnstakeWatcherFeeOwner:        DefaultParamsOwner.Address().String(),
		MessagePauseWatcherFeeOwner:          DefaultParamsOwner.Address().String(),
		MessageUnpauseWatcherFeeOwner:        DefaultParamsOwner.Address().String(),
		MessageWatcherPauseServicerFeeOwner:  DefaultParamsOwner.Address().String(),
		MessageTestScoreFeeOwner:             DefaultParamsOwner.Address().String(),
		MessageProveTestScoreFeeOwner:        DefaultParamsOwner.Address().String(),
		MessageStakeAppFeeOwner:              DefaultParamsOwner.Address().String(),
		MessageEditStakeAppFeeOwner:          DefaultParamsOwner.Address().String(),
		MessageUnstakeAppFeeOwner:            DefaultParamsOwner.Address().String(),
		MessagePauseAppFeeOwner:              DefaultParamsOwner.Address().String(),
		MessageUnpauseAppFeeOwner:            DefaultParamsOwner.Address().String(),
		MessageStakeValidatorFeeOwner:        DefaultParamsOwner.Address().String(),
		MessageEditStakeValidatorFeeOwner:    DefaultParamsOwner.Address().String(),
		MessageUnstakeValidatorFeeOwner:      DefaultParamsOwner.Address().String(),
		MessagePauseValidatorFeeOwner:        DefaultParamsOwner.Address().String(),
		MessageUnpauseValidatorFeeOwner:      DefaultParamsOwner.Address().String(),
		MessageStakeServicerFeeOwner:         DefaultParamsOwner.Address().String(),
		MessageEditStakeServicerFeeOwner:     DefaultParamsOwner.Address().String(),
		MessageUnstakeServicerFeeOwner:       DefaultParamsOwner.Address().String(),
		MessagePauseServicerFeeOwner:         DefaultParamsOwner.Address().String(),
		MessageUnpauseServicerFeeOwner:       DefaultParamsOwner.Address().String(),
		MessageChangeParameterFeeOwner:       DefaultParamsOwner.Address().String(),
	}
}
