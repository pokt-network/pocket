package main

import (
	"github.com/pokt-network/pocket/app/client/cli"
	"github.com/pokt-network/pocket/logger"

	"github.com/spf13/cobra/doc"
)

func main() {
	cmd := cli.GetRootCmd()
	err := doc.GenMarkdownTree(cmd, "../doc/commands")
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("failed to generate markdown tree")
	}
}
