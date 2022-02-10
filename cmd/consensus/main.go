package main

import (
	"flag"
	"log"

	"pocket/consensus/pkg/config"
	"pocket/consensus/pkg/pocket"
	"pocket/shared/context"
)

func main() {
	config_filename := flag.String("config", "", "Relative or absolute path to config file.")
	flag.Parse()

	ctx := context.EmptyPocketContext()
	cfg := config.LoadConfig(*config_filename)

	pocketNode, err := pocket.Create(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create pocket node: %s", err)
	}

	if err = pocketNode.Start(ctx); err != nil {
		log.Fatalf("Failed to start pocket node: %s", err)
	}
}
