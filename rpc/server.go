package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

type rpcServer struct {
	base_modules.IntegrableModule

	logger modules.Logger
}

var (
	_ ServerInterface          = &rpcServer{}
	_ modules.IntegrableModule = &rpcServer{}

	errInvalidJsonRpc = errors.New("JSONRPC validation failed")
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

// Validate returns an error if the payload struct is not valid JSONRPC
func (p *JSONRPCPayload) Validate() error {
	if p.Method == "" {
		return fmt.Errorf("%w: missing method field", errInvalidJsonRpc)
	}

	if p.Jsonrpc != "2.0" {
		return fmt.Errorf("%w: invalid JSONRPC field value: %q", errInvalidJsonRpc, p.Jsonrpc)
	}

	return nil
}

// UnmarshalJSON is the custom unmarshaller for JsonRpcId type. It is needed because JSONRPC spec allows the "id" field to be nil, an integer, or a string.
//
//	See the following link for more details:
//	https://www.jsonrpc.org/specification#request_object
func (i *JsonRpcId) UnmarshalJSON(data []byte) error {
	var v int64
	if err := json.Unmarshal(data, &v); err == nil {
		i.Id = []byte(fmt.Sprintf("%d", v))
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		i.Id = []byte(s)
		return nil
	}

	return fmt.Errorf("invalid JSONRPC ID value: %v", data)
}
