package test_artifacts

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/converters"
)

var (
	DefaultChains              = []string{"0001"}
	DefaultServiceURL          = ""
	DefaultStakeAmount         = big.NewInt(1000000000000)
	DefaultStakeAmountString   = converters.BigIntToString(DefaultStakeAmount)
	DefaultMaxRelays           = big.NewInt(1000000)
	DefaultMaxRelaysString     = converters.BigIntToString(DefaultMaxRelays)
	DefaultAccountAmount       = big.NewInt(100000000000000)
	DefaultAccountAmountString = converters.BigIntToString(DefaultAccountAmount)
	DefaultPauseHeight         = int64(-1)
	DefaultUnstakingHeight     = int64(-1)
	DefaultChainID             = "testnet"
	DefaultServiceUrl          = "https://localhost.com"
	ServiceUrlFormat           = "node%d.consensus:8080"
	DefaultMaxBlockBytes       = uint64(4000000)
)
