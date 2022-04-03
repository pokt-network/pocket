package p2p

import (
	"fmt"
)

// TOOO(derrandz): add/implement (or consolidate with shared/types/error.go) error codes for the following:
// const (
// 	ErrMsgTooLong = iota
// 	ErrMsgTooShort
// 	ErrInternal
// 	ErrConnClosed
// 	ErrConnTimeout
// )

var (
	ErrMissingRequiredArg = func(value string) error {
		return fmt.Errorf("socket error: missing required argument: %s", value)
	}
	ErrSocketRequestTimedOut = func(addr string, nonce uint32) error {
		return fmt.Errorf("socket error: request timed out while waiting on ACK. nonce=%d, addr=%s", nonce, addr)
	}
	ErrSocketUndefinedKind = func(kind string) error {
		return fmt.Errorf("socket error: undefined given socket kind: %s", kind)
	}
	ErrPeerHangUp = func(err error) error {
		return fmt.Errorf("socket error: Peer hang up: %s", err.Error())
	}
	ErrUnexpected = func(err error) error {
		return fmt.Errorf("socket error: Unexpected peer error: %s", err.Error())
	}
	ErrPayloadTooBig = func(bodyLength, acceptedLength uint) error {
		return fmt.Errorf("socket error: cannot read a buffer of length %d, the accepted body length is %d", bodyLength, acceptedLength)
	}
	ErrSocketIOStartFailed = func(socketType string) error {
		return fmt.Errorf("socket error: (%s socket) failed to start socket reading/writing (io)", socketType)
	}
)
