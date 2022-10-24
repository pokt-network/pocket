package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/logger_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go
import (
	"github.com/rs/zerolog"
)

type Logger = zerolog.Logger

type LoggerModule interface {
	Module

	// TODO(#288): Embed `ObservableModule` inside of the `Module` interface so each module has its own context specific logger
	ObservableModule

	// `CreateLoggerForModule` returns the logger with additional context (e.g. module name)
	// https://github.com/rs/zerolog#sub-loggers-let-you-chain-loggers-with-additional-context
	CreateLoggerForModule(string) Logger
}
