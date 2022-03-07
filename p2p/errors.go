package p2p

import (
	"errors"
	"fmt"
	"io"
	"net"
)

const P2PModErrPrefix = "[p2p error]:"

var (
	ErrNotCreated = errors.New("Module error: P2P Module not created. Trying to start the p2p module before calling create.")

	ErrPeerHangUp func(error) error = func(err error) error {
		strerr := fmt.Sprintf("%s Peer Hang Up Error: %s", P2PModErrPrefix, err.Error())
		return errors.New(strerr)
	}
	ErrUnexpected func(error) error = func(err error) error {
		strerr := fmt.Sprintf("%s Unexpected Peer Error: %s", P2PModErrPrefix, err.Error())
		return errors.New(strerr)
	}
	ErrMissingOrEmptyConfigField func(string) error = func(name string) error {
		strerr := fmt.Sprintf("%s Missing or empty required configuration field: %s", P2PModErrPrefix, name)
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
