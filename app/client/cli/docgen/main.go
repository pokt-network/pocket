package main

import (
	"log"

	"github.com/pokt-network/pocket/app/client/cli"

	"github.com/spf13/cobra/doc"
)

func main() {
	cmd := cli.GetRootCmd()
	err := doc.GenMarkdownTree(cmd, "../doc/commands")
	if err != nil {
		log.Fatal(err)
	}
}
