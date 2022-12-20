package main

import (
	"flag"
	"log"

	app "github.com/pokt-network/pocket/cmd"
	"github.com/pokt-network/pocket/internal/runtime"
	"github.com/pokt-network/pocket/internal/shared"
)

func main() {
	configFilename := flag.String("config", "", "Relative or absolute path to the config file.")
	genesisFilename := flag.String("genesis", "", "Relative or absolute path to the genesis file.")

	v := flag.Bool("version", false, "")
	flag.Parse()

	if *v {
		log.Printf("Version flag currently unused %s\n", app.AppVersion)
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
