package peerstore_provider

import "fmt"

// TECHDEBT(#519, #556): Standardize how errors are managed throughout the codebase.

type ErrResolvingAddr struct {
	error
}

func NewErrResolvingAddr(err error) *ErrResolvingAddr {
	return &ErrResolvingAddr{error: err}
}

func (e ErrResolvingAddr) Error() string {
	return fmt.Sprintf("error resolving addr: %s", e.error)
}
