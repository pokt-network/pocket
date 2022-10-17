package main

import (
	"flag"
	"log"

	"github.com/pokt-network/pocket/runtime"
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

	runtimeMgr := runtime.NewManagerFromFiles(*configFilename, *genesisFilename)

	pocketNode, err := shared.CreateNode(runtimeMgr)
	if err != nil {
		log.Fatalf("Failed to create pocket node: %s", err)
	}

	if err = pocketNode.Start(); err != nil {
		log.Fatalf("Failed to start pocket node: %s", err)
	}
}
