package helpers

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pokt-network/pocket/app/client/cli/flags"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p"
	rpcCHP "github.com/pokt-network/pocket/p2p/providers/current_height_provider/rpc"
	rpcPSP "github.com/pokt-network/pocket/p2p/providers/peerstore_provider/rpc"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared/modules"
)

// P2PDependenciesPreRunE initializes peerstore & current height providers, and a
// p2p module which consumes them. Everything is registered to the bus.
func P2PDependenciesPreRunE(cmd *cobra.Command, _ []string) error {
	// TECHDEBT: this is to keep backwards compatibility with localnet
	flags.ConfigPath = runtime.GetEnv("CONFIG_PATH", "build/config/config.validator1.json")

	runtimeMgr := runtime.NewManagerFromFiles(
		flags.ConfigPath, genesisPath,
		runtime.WithClientDebugMode(),
		runtime.WithRandomPK(),
	)

	bus := runtimeMgr.GetBus()
	SetValueInCLIContext(cmd, BusCLICtxKey, bus)

	if err := setupPeerstoreProvider(*runtimeMgr, flags.RemoteCLIURL); err != nil {
		return err
	}

	if err := setupRPCCurrentHeightProvider(*runtimeMgr, flags.RemoteCLIURL); err != nil {
		return err
	}

	setupAndStartP2PModule(*runtimeMgr)

	return nil
}

func setupPeerstoreProvider(rm runtime.Manager, rpcURL string) error {
	// Ensure `PeerstoreProvider` exists in the modules registry.
	if _, err := rpcPSP.Create(rm.GetBus(), rpcPSP.WithCustomRPCURL(rpcURL)); err != nil {
		return err
	}
	return nil
}

func setupRPCCurrentHeightProvider(rm runtime.Manager, rpcURL string) error {
	// Ensure `CurrentHeightProvider` exists in the modules registry.
	_, err := rpcCHP.Create(
		rm.GetBus(),
		rpcCHP.WithCustomRPCURL(rpcURL),
	)
	if err != nil {
		return fmt.Errorf("setting up current height provider: %w", err)
	}
	return nil
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
