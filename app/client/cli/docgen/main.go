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
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("failed to get working directory")
	}

	docsPath, err := filepath.Abs(workingDir + "/../../doc/commands")
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("failed to get absolute path")
	}

	cmd := cli.GetRootCmd()
	err = doc.GenMarkdownTree(cmd, docsPath)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("failed to generate markdown tree")
	}
}
