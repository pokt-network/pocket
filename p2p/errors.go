package p2p

import (
	"errors"
	stdio "io"
	"net"
)

func isErrEOF(err error) bool {
	if errors.Is(err, stdio.EOF) {
		return true
	}

	var netErr *net.OpError

	if errors.As(err, &netErr) && netErr.Err.Error() == "use of closed network connection" {
		return true
	}

	return false
}

func isErrTimeout(err error) bool {
	var netErr *net.OpError

	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return false
}
