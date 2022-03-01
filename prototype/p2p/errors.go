package p2p

import (
	"errors"
	"fmt"
	"io"
	"net"
)

var (
	ErrNotCreated                   = errors.New("Module error: P2P Module not created. Trying to start the p2p module before calling create.")
	ErrPeerHangUp func(error) error = func(err error) error {
		strerr := fmt.Sprintf("Peer Hang Up Error: %s", err.Error())
		return errors.New(strerr)
	}
	ErrUnexpected func(error) error = func(err error) error {
		strerr := fmt.Sprintf("Unexpected Peer Error: %s", err.Error())
		return errors.New(strerr)
	}
)

func isErrEOF(err error) bool {
	if errors.Is(err, io.EOF) {
		return true
	}

	var netErr *net.OpError
	return errors.As(err, &netErr) && netErr.Err.Error() == "use of closed network connection"
}

func isErrTimeout(err error) bool {
	var netErr *net.OpError

	return errors.As(err, &netErr) && netErr.Timeout()
}
