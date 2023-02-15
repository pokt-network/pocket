package main

import (
	"os"
	"path/filepath"

	"github.com/pokt-network/pocket/app/client/cli"
	"github.com/pokt-network/pocket/logger"

	"github.com/spf13/cobra/doc"
)

func main() {
	workingDir, err := os.Getwd()
	docsPath, err := filepath.Abs(workingDir + "/../../doc/commands")

	cmd := cli.GetRootCmd()
	err = doc.GenMarkdownTree(cmd, docsPath)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("failed to generate markdown tree")
	}
}
