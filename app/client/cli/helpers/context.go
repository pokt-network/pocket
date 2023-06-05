package helpers

import (
	"context"
	"github.com/spf13/cobra"
)

const BusCLICtxKey cliContextKey = "bus"

// NOTE: this is required by the linter, otherwise a simple string constant would have been enough
type cliContextKey string

func SetValueInCLIContext(cmd *cobra.Command, key cliContextKey, value any) {
	cmd.SetContext(context.WithValue(cmd.Context(), key, value))
}

func GetValueFromCLIContext[T any](cmd *cobra.Command, key cliContextKey) (T, bool) {
	value, ok := cmd.Context().Value(key).(T)
	return value, ok
}
