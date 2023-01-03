package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pokt-network/pocket/app/client/cli"
	"github.com/pokt-network/pocket/logger"
)

func main() {
	ctx := newCLIContext()
	err := cli.ExecuteContext(ctx)
	if ctx.Err() == context.Canceled || err == context.Canceled {
		logger.Global.Logger.Fatal().Msg("aborted")
		return
	}

	if err != nil {
		logger.Global.Logger.Fatal().Err(err).Msg("failed to execute command")
	}
}

func newCLIContext() context.Context {
	var (
		cancelCtx, cancel = context.WithCancel(context.Background())
		quit              = make(chan os.Signal, 1)
	)
	signal.Notify(quit,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		os.Kill, //nolint
		os.Interrupt)
	go func() {
		<-quit
		cancel()
	}()
	return cancelCtx
}
