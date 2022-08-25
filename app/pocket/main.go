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

	pocketNode, err := shared.Create(cfg, genesis)
	if err != nil {
		log.Fatalf("Failed to create pocket node: %s", err)
	}

	if cfg.Rpc.Enabled {
		go rpc.NewRPCServer(pocketNode).StartRPC(cfg.Rpc.Port, cfg.Rpc.Timeout)
	} else {
		log.Println("[WARN] RPC server: OFFLINE")
	}

	if err = pocketNode.Start(); err != nil {
		log.Fatalf("Failed to start pocket node: %s", err)
	}
}
