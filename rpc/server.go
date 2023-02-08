package rpc

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pokt-network/pocket/shared/modules"
)

type rpcServer struct {
	modules.BaseIntegratableModule

	logger modules.Logger
}

var (
	_ ServerInterface            = &rpcServer{}
	_ modules.IntegratableModule = &rpcServer{}
)

func NewRPCServer(bus modules.Bus) *rpcServer {
	s := &rpcServer{}
	s.SetBus(bus)

	return s
}

func (s *rpcServer) StartRPC(port string, timeout uint64, logger *modules.Logger) {
	s.logger = *logger

	s.logger.Info().Msgf("Starting RPC on port " + port)

	e := echo.New()
	middlewares := []echo.MiddlewareFunc{
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogURI:    true,
			LogStatus: true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				s.logger.Info().
					Str("URI", v.URI).
					Int("status", v.Status).
					Msg("request")

				return nil
			},
		}),
		middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Skipper:      middleware.DefaultSkipper,
			ErrorMessage: "Request timed out",
			Timeout:      time.Duration(timeout) * time.Millisecond,
		}),
	}
	if s.GetBus().GetRuntimeMgr().GetConfig().RPC.UseCors {
		s.logger.Info().Msg("Enabling CORS middleware")
		middlewares = append(middlewares, middleware.CORS())
	}
	e.Use(
		middlewares...,
	)

	RegisterHandlers(e, s)

	if err := e.Start(":" + port); err != http.ErrServerClosed {
		s.logger.Fatal().Err(err).Msg("RPC server failed to start")
	}
}
