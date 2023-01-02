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
		logger.Global.Logger.Info().Str("version", app.AppVersion).Msg("Version flag currently unused")
		return
	}

	runtimeMgr := runtime.NewManagerFromFiles(*configFilename, *genesisFilename)

	pocketNode, err := shared.CreateNode(runtimeMgr)
	if err != nil {
		logger.Global.Logger.Fatal().Err(err).Msg("Failed to create pocket node")
	}

	if err = pocketNode.Start(); err != nil {
		logger.Global.Logger.Fatal().Err(err).Msg("Failed to start pocket node")
	}
}
