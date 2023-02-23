package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/logger_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go
import (
	"github.com/rs/zerolog"
)

const LoggerModuleName = "logger"

type Logger = zerolog.Logger

type LoggerModule interface {
	Module

	// TODO(#288): Embed `ObservableModule` inside of the `Module` interface so each module has its own context specific logger
	ObservableModule

	// CreateLoggerForModule returns the logger with additional context (e.g. module name)
	// (see: https://github.com/rs/zerolog#sub-loggers-let-you-chain-loggers-with-additional-context)
	// NB: returns a pointer to mitigate `hugParam` linter error.
	// (see: https://golangci-lint.run/usage/linters/#gocritic)
	CreateLoggerForModule(string) *Logger
}
