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

	v := flag.Bool("version", false, "")
	flag.Parse()

	if *v {
		logger.Global.Info().Str("version", app.AppVersion).Msg("Version flag currently unused")
		return
	}

	runtimeMgr := runtime.NewManagerFromFiles(*configFilename, *genesisFilename)
	bus, err := runtime.CreateBus(runtimeMgr)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to create bus")
	}

	pocketNode, err := shared.CreateNode(bus)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to create pocket node")
	}
	//pocketNode.GetBus().GetConsensusModule().EnableServerMode()

	if err = pocketNode.Start(); err != nil {
		logger.Global.Fatal().Err(err).Msg("Failed to start pocket node")
	}
}
