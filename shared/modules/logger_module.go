package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/logger_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go
import (
	"github.com/rs/zerolog"
)

type LoggerModule interface {
	Module

	// GetLogger returns the logger with additional context (module name)
	// https://github.com/rs/zerolog#sub-loggers-let-you-chain-loggers-with-additional-context
	GetLoggerForModule(string) zerolog.Logger
}
