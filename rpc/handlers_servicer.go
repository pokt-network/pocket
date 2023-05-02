package rpc

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *rpcServer) PostV1ServicerRelay(ctx echo.Context) error {

	// TODO: Move to the servicer module
	servicerModuleName := "servicer"
	found := false
	for _, am := range s.GetBus().GetUtilityModule().GetActorModules() {
		if am.GetModuleName() == servicerModuleName {
			found = true
			break
		}
	}

	if !found {
		return ctx.String(http.StatusInternalServerError, "node is unable to serve relays")
	}

	return nil
}
