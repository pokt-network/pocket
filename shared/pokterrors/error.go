package pokterrors

// DISCUSS(M5): Evaluate how Pocket specific errors should be managed and returned to the client
// TECHDEBT: Remove reference to the term `Proto`; it's why we created a codec package

import (
	"errors"
	"fmt"

	"github.com/pokt-network/pocket/shared/modules"
)

type Code int32

type Error interface {
	Code() Code
	error
}

var _ Error = &stdErr{}

type stdErr struct {
	CodeError Code
	module    string
	error
}

func (se *stdErr) Error() string {
	return fmt.Sprintf("CODE: %v-%v, ERROR: %s", se.Module(), se.Code(), se.error.Error())
}

func (se *stdErr) Code() Code {
	return se.CodeError
}

func (se *stdErr) Module() string {
	return se.module
}

func NewUtilityError(code UtilityErrorCode, msg string) Error {
	return &stdErr{
		CodeError: Code(code),
		module:    modules.UtilityModuleName,
		error:     errors.New(msg),
	}
}
