package p2p

import (
	"time"

	"github.com/pokt-network/pocket/p2p/types"
)

func (m *p2pModule) ping(addr string) (bool, error) { // TODO(derrandz): What's the proper return instead of bool for ping/pong sequences?
	// TODO(derrandz): refactor this to use types.P2PMessage
	var pongbytes []byte

	pingbytes := []byte(types.PING)

	timedout := make(chan int)
	ponged := make(chan int)
	errored := make(chan error)

	go func() {
		<-time.After(time.Millisecond * 500)
		timedout <- 1
	}()

	go func() {
		response, err := m.request(addr, pingbytes, true)

		if err != nil {
			errored <- err
		}

		pongbytes = response
		ponged <- 1
	}()

	select {

	case <-timedout:
		return false, nil

	case err := <-errored:
		return false, err

	case <-ponged:
		pong, err := m.c.Decode(pongbytes)
		pongbytes := pong.([]byte)

		if err != nil {
			return false, err
		}

		pongmsg := string(pongbytes)

		if pongmsg != types.PONG {
			return false, nil
		}

		return true, nil
	}
} // TODO(derrandz): should we use UDP requests for ping?

// TODO(derrandz): test
func (m *p2pModule) pong(nonce uint32, sequence []byte, source string) error {
	if nonce != 0 && string(sequence) == types.PING {
		pongbytes, err := m.c.Encode([]byte(types.PONG))

		if err != nil {
			return err
		}

		err = m.respond(uint32(nonce), false, source, pongbytes, false)

		if err != nil {
			return err
		}
	}
	return nil
}
