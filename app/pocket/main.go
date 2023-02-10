package main

import (
	"flag"

	"github.com/pokt-network/pocket/app"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared"
)

func main() {
	configFilename := flag.String("config", "", "Relative or absolute path to the config file.")
	genesisFilename := flag.String("genesis", "", "Relative or absolute path to the genesis file.")
	bootstrapNodes := flag.String("bootstrap-nodes", "", "Comma separated list of bootstrap nodes.")

	v := flag.Bool("version", false, "")
	flag.Parse()

	if *v {
		logger.Global.Info().Str("version", app.AppVersion).Msg("Version flag currently unused")
		return
	}

	options := []func(*runtime.Manager){}
	if bootstrapNodes != nil && *bootstrapNodes != "" {
		options = append(options, runtime.WithCustomBootstrapNodes(*bootstrapNodes))
	}

	runtimeMgr := runtime.NewManagerFromFiles(*configFilename, *genesisFilename, options...)
	bus, err := runtime.CreateBus(runtimeMgr)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to create bus")
	}

	pocketNode, err := shared.CreateNode(bus)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to create pocket node")
	}
	pocketNode.GetBus().GetConsensusModule().EnableServerMode()

	if err = pocketNode.Start(); err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to start pocket node")
	}
}
