package rpc

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pokt-network/pocket/shared"
	"github.com/pokt-network/pocket/shared/config"
)

type rpcServer struct {
	node *shared.Node
}

var _ ServerInterface = &rpcServer{}

func NewRPCServer(pocketNode *shared.Node) *rpcServer {
	return &rpcServer{
		node: pocketNode,
	}
}

func (s *rpcServer) StartRPC(port string, timeout uint64) {
	log.Printf("Starting RPC on port %s...\n", port)

	e := echo.New()
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "Request timed out",
		Timeout:      config.DefaultRPCTimeout * time.Millisecond,
	}))
	RegisterHandlers(e, s)

	if err := e.Start(":" + port); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
