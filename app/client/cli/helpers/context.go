package helpers

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pokt-network/pocket/shared/modules"
)

const BusCLICtxKey cliContextKey = "bus"

var ErrCxtFromBus = fmt.Errorf("could not get context from bus")

// NOTE: this is required by the linter, otherwise a simple string constant would have been enough
type cliContextKey string

func SetValueInCLIContext(cmd *cobra.Command, key cliContextKey, value any) {
	cmd.SetContext(context.WithValue(cmd.Context(), key, value))
}

func GetValueFromCLIContext[T any](cmd *cobra.Command, key cliContextKey) (T, bool) {
	value, ok := cmd.Context().Value(key).(T)
	return value, ok
}

func GetBusFromCmd(cmd *cobra.Command) (modules.Bus, error) {
	bus, ok := GetValueFromCLIContext[modules.Bus](cmd, BusCLICtxKey)
	if !ok {
		return nil, ErrCxtFromBus
	}

	return bus, nil
}
