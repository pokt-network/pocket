package peerstore_provider

import "fmt"

type ErrResolvingAddr struct {
	error
}

func NewErrResolvingAddr(err error) *ErrResolvingAddr {
	return &ErrResolvingAddr{error: err}
}

func (e ErrResolvingAddr) Error() string {
	return fmt.Sprintf("error resolving addr: %s", e.error)
}
