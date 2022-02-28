package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/pokt-network/pocket/shared"
	"github.com/pokt-network/pocket/shared/config"
)

// TODO(iajrz): Document where/how the version number is injected from the build process into this variable.
var version = "UNKNOWN"

func main() {
	config_filename := flag.String("config", "", "Relative or absolute path to config file.")
	version := flag.Bool("version", false, "")
	flag.Parse()

	if *version {
		// TODO(iajrz): Fix/remove how version is injected into this variable and its type.
		fmt.Printf("Version: %b\n", version)
	}

	cfg := config.LoadConfig(*config_filename)

	pocketNode, err := shared.Create(cfg)
	if err != nil {
		log.Fatalf("Failed to create pocket node: %s", err)
	}

	if err = pocketNode.Start(); err != nil {
		log.Fatalf("Failed to start pocket node: %s", err)
	}
}
