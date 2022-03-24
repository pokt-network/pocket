package p2p

import (
	"fmt"
)

var (
	ErrSocketEmptyContextValue = func(value string) error {
		return fmt.Errorf("socket error: the provided context does not have the value: %s", value)
	}
	ErrSocketRequestTimedOut = func(addr string, nonce uint32) error {
		return fmt.Errorf("socket error: request timedout while waiting on ACK. nonce=%d, addr=%s", nonce, addr)
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
		return fmt.Errorf("socket error: cannot read a buffer of length %d, the accepted body length is %d.", bodyLength, acceptedLength)
	}
)
