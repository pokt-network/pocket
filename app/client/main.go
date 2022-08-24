package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pokt-network/pocket/app/client/cli"
)

func main() {
	ctx := newCLIContext()
	err := cli.ExecuteContext(ctx)
	if ctx.Err() == context.Canceled || err == context.Canceled {
		fmt.Println("aborted")
		return
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
