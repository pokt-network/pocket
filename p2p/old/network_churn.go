package p2p

import (
	"time"

	"github.com/pokt-network/pocket/p2p/types"
)

// Sends a small PING byte to the provided address
// the only encoding that takes place is by the socket's wireCodec, using BigEndian Binary bytes representation
// This is intentionally meant to be small
// TODO(derrandz): What's the proper return instead of bool for ping/pong sequences?
// TODO(derrandz): should we use UDP requests for ping?
func (m *p2pModule) ping(addr string) (bool, error) {
	pingbytes := []byte(types.PING)

	timedout := make(chan int)
	ponged := make(chan []byte)
	errored := make(chan error)

	//TODO(derrandz): Remove in favor of using timeouts on the socket read process itself
	go func() {
		<-time.After(time.Millisecond * 1000)
		timedout <- 1
	}()

	go func() {
		// TODO(derrandz): maybe replace the meta properties such as encoded=true|false by a haltable context with values
		response, err := m.request(addr, pingbytes, false)

		if err != nil {
			errored <- err
		}

		ponged <- response
	}()

	select {
	case <-timedout:
		return false, nil

	case err := <-errored:
		return false, err

	case bytes := <-ponged:
		pongmsg := string(bytes)

		if pongmsg != types.PONG {
			return false, nil
		}

		return true, nil
	}
}

// TODO(derrandz): should we use UDP requests for ping?
func (m *p2pModule) pong(nonce uint32, sequence []byte, source string) error {
	if nonce != 0 && string(sequence) == types.PING {
		pongbytes := []byte(types.PONG)
		err := m.respond(uint32(nonce), false, source, pongbytes, false)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
