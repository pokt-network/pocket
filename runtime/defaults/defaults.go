package defaults

import (
	"fmt"
	"math/big"

	"github.com/pokt-network/pocket/shared/converters"
)

const (
	defaultRPCPort    = "50832"
	defaultRPCHost    = "localhost"
	defaultRPCTimeout = 30000
)

// TODO: some of these defaults are used for production while others are used for testing. Need to separate them.
var (
	DefaultChains                  = []string{"0001"}
	DefaultServiceURL              = ""
	DefaultStakeAmount             = big.NewInt(1000000000000)
	DefaultStakeAmountString       = converters.BigIntToString(DefaultStakeAmount)
	DefaultMaxRelays               = big.NewInt(1000000)
	DefaultMaxRelaysString         = converters.BigIntToString(DefaultMaxRelays)
	DefaultAccountAmount           = big.NewInt(100000000000000)
	DefaultAccountAmountString     = converters.BigIntToString(DefaultAccountAmount)
	DefaultPauseHeight             = int64(-1)
	DefaultUnstakingHeight         = int64(-1)
	DefaultChainID                 = "testnet"
	DefaultMaxBlockBytes           = uint64(4000000)
	ServiceUrlFormat               = "node%d.consensus:8080"
	DefaultRpcPort                 = defaultRPCPort
	DefaultRpcTimeout              = uint64(defaultRPCTimeout)
	DefaultRemoteCLIURL            = fmt.Sprintf("http://%s:%s", defaultRPCHost, defaultRPCPort)
	DefaultP2PMempoolMaxNonceCount = uint64(1e6)
)
