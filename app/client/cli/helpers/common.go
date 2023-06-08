package helpers

import (
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	genesisPath = runtime.GetEnv("GENESIS_PATH", "build/config/genesis.json")
	RpcHost     string

	// P2PMod is initialized in order to broadcast a message to the local network
	P2PMod modules.P2PModule
)
