package defaults

import (
	"fmt"
)

const (
	defaultRPCPort    = "50832"
	defaultRPCHost    = "localhost"
	defaultRPCTimeout = 30000
)

var (
	DefaultRemoteCLIURL       = fmt.Sprintf("http://%s:%s", defaultRPCHost, defaultRPCPort)
	DefaultRpcPort            = defaultRPCPort
	DefaultRpcTimeout         = uint64(defaultRPCTimeout)
	DefaultP2PMaxMempoolCount = uint64(1e5)
)
