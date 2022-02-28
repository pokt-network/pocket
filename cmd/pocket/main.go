package main

import (
	"flag"
	"log"
	"pocket/shared"
	"pocket/shared/config"
)

// TODO(iajrz): Do we need this default variable?
// var version = "UNKNOWN"

func main() {
	config_filename := flag.String("config", "", "Relative or absolute path to config file.")
	version := flag.Bool("version", false, "")
	flag.Parse()

	if *version {
		log.Printf("Version: %b\n", version)
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
