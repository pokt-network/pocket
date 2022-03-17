package p2p

import (
	"fmt"
)

var (
	ErrSocketEmptyContextValue func(string) error = func(value string) error {
		return fmt.Errorf("socket error: the provided context does not have the value: %s", value)
	}
	ErrSocketRequestTimedOut func(string, uint32) error = func(addr string, nonce uint32) error {
		return fmt.Errorf("socket error: request timedout while waiting on ACK. nonce=%d, addr=%s", nonce, addr)
	}
	ErrSocketUndefinedKind func(string) error = func(kind string) error {
		return fmt.Errorf("socket error: undefined given socket kind: %s", kind)
	}
	ErrPeerHangUp func(error) error = func(err error) error {
		return fmt.Errorf("socket error: Peer hang up: %s", err.Error())
	}
	ErrUnexpected func(error) error = func(err error) error {
		return fmt.Errorf("socket error: Unexpected peer error: %s", err.Error())
	}
)
