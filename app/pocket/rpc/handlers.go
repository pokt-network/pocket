package rpc

import (
	"encoding/hex"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/app"
)

func (s *rpcServer) GetV1Health(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}

func (s *rpcServer) GetV1Version(ctx echo.Context) error {
	return ctx.String(http.StatusOK, app.AppVersion)
}

func (s *rpcServer) PostV1ClientBroadcastTxSync(ctx echo.Context) error {
	params := new(SendRawTxParams)
	if err := ctx.Bind(params); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}
	bz, err := hex.DecodeString(params.RawHexBytes)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "cannot decode")
	}
	bus := s.node.GetBus()
	height := bus.GetConsensusModule().CurrentHeight()
	uCtx, err := bus.GetUtilityModule().NewContext(int64(height))
	if err != nil {
		defer func() { log.Fatalf("[ERROR] Failed to create UtilityContext: %v", err) }()
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	err = uCtx.CheckTransaction(bz)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}
