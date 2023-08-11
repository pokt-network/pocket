package utils

import (
	"context"
	"net/http"

	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/shared/modules"
)

func CheckHealth(ctx context.Context, serviceURL string, logger *modules.Logger) error {
	client, err := rpc.NewClientWithResponses(serviceURL)
	if err != nil {
		logger.Error().Err(err).
			Str("serviceURL", serviceURL).
			Msg("error creating RPC client")
		return err
	}
	healthCheck, err := client.GetV1Health(ctx)
	if err != nil || healthCheck == nil || healthCheck.StatusCode != http.StatusOK {
		logger.Error().Err(err).
			Str("serviceURL", serviceURL).
			Msg("error getting a green health check from bootstrap node")
		return err
	}
	return nil
}
