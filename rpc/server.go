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
var _ modules.IntegratableModule = &rpcServer{}

func NewRPCServer(bus modules.Bus) *rpcServer {
	s := &rpcServer{}
	s.SetBus(bus)
	return s
}

func (s *rpcServer) StartRPC(port string, timeout uint64) {
	log.Printf("Starting RPC on port %s...\n", port)

	e := echo.New()
	middlewares := []echo.MiddlewareFunc{
		middleware.Logger(),
		middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Skipper:      middleware.DefaultSkipper,
			ErrorMessage: "Request timed out",
			Timeout:      time.Duration(defaults.DefaultRpcTimeout) * time.Millisecond,
		}),
	}
	if s.bus.GetRuntimeMgr().GetConfig().GetRPCConfig().GetUseCors() {
		log.Println("Enabling CORS middleware")
		middlewares = append(middlewares, middleware.CORS())
	}
	e.Use(
		middlewares...,
	)

	RegisterHandlers(e, s)

	if err := e.Start(":" + port); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func (s *rpcServer) SetBus(bus modules.Bus) {
	s.bus = bus
}

func (s *rpcServer) GetBus() modules.Bus {
	return s.bus
}
