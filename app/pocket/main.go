package main

import (
	"flag"
	"log"

	"github.com/pokt-network/pocket/app/client/rpc"
	"github.com/pokt-network/pocket/shared"
	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
)

// See `docs/build/README.md` for details on how this is injected via mage.
var version = "UNKNOWN"

func main() {
	configFilename := flag.String("config", "", "Relative or absolute path to the config file.")
	genesisFilename := flag.String("genesis", "", "Relative or absolute path to the genesis file.")

	v := flag.Bool("version", false, "")
	flag.Parse()

	if *v {
		log.Printf("Version flag currently unused %s\n", version)
		return
	}

	cfg, genesis := test_artifacts.ReadConfigAndGenesisFiles(*configFilename, *genesisFilename)
	// DISCUSS(deblasis): poor man's feature flag: no port => no RPC
	if cfg.RPC.Enable {
		go rpc.StartRPC(cfg.RPC.Port, cfg.RPC.Timeout)
	} else {
		log.Println("RPC server: OFFLINE")
	}

	pocketNode, err := shared.Create(cfg, genesis)
	if err != nil {
		log.Fatalf("Failed to create pocket node: %s", err)
	}

	if err = pocketNode.Start(); err != nil {
		log.Fatalf("Failed to start pocket node: %s", err)
	}

}
