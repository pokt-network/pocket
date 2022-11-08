package main

import (
	"context"
	"github.com/pokt-network/pocket/app/client/cli"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := newCLIContext()
	err := cli.ExecuteContext(ctx)
	if ctx.Err() == context.Canceled || err == context.Canceled {
		log.Fatalf("aborted\n")
		return
	}

	if err != nil {
		log.Fatalf("err: %v\n", err)
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
