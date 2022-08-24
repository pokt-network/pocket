package main

import (
	"flag"
	"log"

	"github.com/pokt-network/pocket/app/client/rpc"
	"github.com/pokt-network/pocket/shared"
	"github.com/pokt-network/pocket/shared/config"
)

// See `docs/build/README.md` for details on how this is injected via mage.
var version = "UNKNOWN"

func main() {
	config_filename := flag.String("config", "", "Relative or absolute path to config file.")
	v := flag.Bool("version", false, "")
	flag.Parse()

	if *v {
		log.Printf("Version flag currently unused %s\n", version)
		return
	}

	cfg := config.LoadConfig(*config_filename)
	if cfg.RPCConfig.Enable {
		go rpc.StartRPC(cfg.RPCConfig.Port, cfg.RPCConfig.Timeout)
	} else {
		log.Println("[WARN] RPC server: OFFLINE")
	}

	pocketNode, err := shared.Create(cfg)
	if err != nil {
		log.Fatalf("Failed to create pocket node: %s", err)
	}

	if err = pocketNode.Start(); err != nil {
		log.Fatalf("Failed to start pocket node: %s", err)
	}
}
