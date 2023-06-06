package helpers

import (
	"fmt"

	"github.com/pokt-network/pocket/app/client/cli/flags"
	"github.com/spf13/cobra"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p"
	rpc2 "github.com/pokt-network/pocket/p2p/providers/current_height_provider/rpc"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider/rpc"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/modules"
)

// P2PDependenciesPreRunE initializes peerstore & current height providers, and a
// p2p module with consumes them. Everything is registered to the bus.
func P2PDependenciesPreRunE(cmd *cobra.Command, _ []string) error {
	// TECHDEBT: this is to keep backwards compatibility with localnet
	flags.ConfigPath = runtime.GetEnv("CONFIG_PATH", "build/config/config1.json")
	rpcURL := fmt.Sprintf("http://%s:%s", RpcHost, defaults.DefaultRPCPort)

	runtimeMgr := runtime.NewManagerFromFiles(
		flags.ConfigPath, GenesisPath,
		runtime.WithClientDebugMode(),
		runtime.WithRandomPK(),
	)

	bus := runtimeMgr.GetBus()
	SetValueInCLIContext(cmd, BusCLICtxKey, bus)

	setupPeerstoreProvider(*runtimeMgr, rpcURL)
	setupCurrentHeightProvider(*runtimeMgr, rpcURL)
	setupAndStartP2PModule(*runtimeMgr)

	return nil
}

func setupPeerstoreProvider(rm runtime.Manager, rpcURL string) {
	bus := rm.GetBus()
	modulesRegistry := bus.GetModulesRegistry()
	pstoreProvider := rpc.NewRPCPeerstoreProvider(
		rpc.WithP2PConfig(rm.GetConfig().P2P),
		rpc.WithCustomRPCURL(rpcURL),
	)
	modulesRegistry.RegisterModule(pstoreProvider)
}

func setupCurrentHeightProvider(rm runtime.Manager, rpcURL string) {
	bus := rm.GetBus()
	modulesRegistry := bus.GetModulesRegistry()
	currentHeightProvider := rpc2.NewRPCCurrentHeightProvider(
		rpc2.WithCustomRPCURL(rpcURL),
	)
	modulesRegistry.RegisterModule(currentHeightProvider)
}

func setupAndStartP2PModule(rm runtime.Manager) {
	bus := rm.GetBus()
	mod, err := p2p.Create(bus)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to create p2p module")
	}

	var ok bool
	P2PMod, ok = mod.(modules.P2PModule)
	if !ok {
		logger.Global.Fatal().Msgf("unexpected P2P module type: %T", mod)
	}

	if err := P2PMod.Start(); err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to start p2p module")
	}
}
