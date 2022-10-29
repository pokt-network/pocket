package rpc

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/modules"
)

type rpcServer struct {
	bus modules.Bus
}

var _ ServerInterface = &rpcServer{}

func NewRPCServer(bus modules.Bus) *rpcServer {
	return &rpcServer{
		bus: bus,
	}
}

func (s *rpcServer) StartRPC(port string, timeout uint64) {
	log.Printf("Starting RPC on port %s...\n", port)

	e := echo.New()
	e.Use(
		middleware.Logger(),
		middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Skipper:      middleware.DefaultSkipper,
			ErrorMessage: "Request timed out",
			Timeout:      time.Duration(defaults.DefaultRpcTimeout) * time.Millisecond,
		}),
	)
	RegisterHandlers(e, s)

	if err := e.Start(":" + port); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
