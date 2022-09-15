package main

import (
	"flag"
	"log"

	"github.com/pokt-network/pocket/shared"
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
	pocketNode, err := shared.Create(*configFilename, *genesisFilename)
	if err != nil {
		log.Fatalf("Failed to create pocket node: %s", err)
	}

	// TECHDEBT: improve configuration handling. There's no way for us to access config at this point. We woudn't want to deserialize/map to a struct here, we should get a typed structure to begin with. This is not Javascript :)
	// Also, I have noticed that the `pocketNode.GetBus().GetConfig()` call is a null pointer operation in LocalNet
	//
	// Because of the above, RPC server is disabled for now
	//
	// if cfg.Rpc.Enabled {
	// 	go rpc.NewRPCServer(pocketNode).StartRPC(cfg.Rpc.Port, cfg.Rpc.Timeout)
	// } else {
	log.Println("[WARN] RPC server: OFFLINE")
	// }

	if err = pocketNode.Start(); err != nil {
		log.Fatalf("Failed to start pocket node: %s", err)
	}
}
