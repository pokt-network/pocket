package main

import (
	"flag"
	"log"

	"pocket/consensus/pkg/config"
	"pocket/consensus/pkg/pocket"
	"pocket/consensus/pkg/shared/context"
)

func main() {
	config_filename := flag.String("config", "", "Relative or absolute path to config file.")
	flag.Parse()

	log.Println("OLSH", config_filename)

	context := context.EmptyPocketContext()
	config := config.LoadConfig(*config_filename)

	pocketNode, err := pocket.Create(context, config)
	if err != nil {
		log.Fatalf("Failed to create pocket node: %s", err)
	}

	if err = pocketNode.Start(context); err != nil {
		log.Fatalf("Failed to start pocket node: %s", err)
	}
}
