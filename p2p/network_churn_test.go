package p2p

import (
	"fmt"
	"sync"
	"testing"

	"github.com/pokt-network/pocket/p2p/types"
	shared "github.com/pokt-network/pocket/shared/config"
	"github.com/stretchr/testify/assert"
)

func TestNetworkChurn_Ping(t *testing.T) {
	config := &shared.P2PConfig{
		Protocol:         "tcp",
		Address:          []byte("0.0.0.0:4032"),
		ExternalIp:       "0.0.0.0:4032",
		Peers:            []string{"0.0.0.0:30303"},
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       1024 * 4,
		WireHeaderLength: 9,
		TimeoutInMs:      200,
	}

	m := newP2PModule()

	err := m.initialize(config)

	assert.Nil(
		t,
		err,
		fmt.Sprintf("Error: failed to initialize got: %s", err),
	)

	t.Log("Ping: p2p module initialize")

	go m.listen()
	_, waiting := <-m.ready

	assert.Equal(
		t,
		waiting,
		false,
		"Request error: peer still not started after Listen",
	)

	t.Log("Ping: peer started listening...")

	addr := "127.0.0.1:2313"
	ready, _, data, respond := ListenAndServe(addr, int(m.config.BufferSize), 200)

	select {
	case v := <-ready:
		assert.NotEqual(
			t,
			v,
			0,
			"Ping: Encoutered error while starting the mock peer",
		)
	}

	t.Log("Ping: mock peer listening...")

	wg := &sync.WaitGroup{}
	errors := make(chan error)
	responses := make(chan bool)

	wg.Add(1)
	go func() {
		t.Log("Ping: pinging the mock peer...")
		alive, err := m.ping(addr)
		if err != nil {
			errors <- err
			close(responses)
			assert.Fail(t, fmt.Sprintf("Ping: pinging the mock peer failed. %s", err))
		}
		responses <- alive
		t.Logf("Ping: pinging the mock peer succeeded.")

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		c := newWireCodec()
		select {
		case d := <-data:
			{
				t.Logf("Ping: mock peer received the ping message successfully")

				nonce, encoding, buff, _, err := c.decode(d.buff)
				assert.Nilf(
					t,
					err,
					"Ping: mock peer encountered error while decoding received ping sequence: %s", err,
				)

				sequence := string(buff)
				assert.Equalf(
					t,
					sequence,
					types.PING,
					"Ping: mock peer expectes to receive ping sequence, m.t %s instead",
					sequence,
				)

				t.Log("Ping: mock peer is preparing to pong.")
				respond <- c.encode(encoding, false, nonce, []byte(types.PONG), false)
				t.Log("Ping: mock peer sent the pong sequence back.")
				wg.Done()
			}
		}
	}()

	wg.Add(1)
	go func() {
		select {
		case err := <-errors:
			t.Errorf("Ping: mock peer received error while waiting for ping: %s", err.Error())

		case alive, open := <-responses:
			t.Logf("Ping: mock peer responded..")
			assert.Equalf(
				t,
				open,
				true,
				"Ping: p2p peer encountered error while waiting to receive a pong response",
			)

			assert.Equalf(
				t,
				alive,
				true,
				"Ping: p2p peer expected the mock peer to be alive and recieve a proper pong sequence, got the following instead: alive=%v",
				alive,
			)
			t.Logf("Ping: Success")
		}
		wg.Done()
	}()

	wg.Wait()

}

func TestNetworkChurn_Pong(t *testing.T) {
	config := &shared.P2PConfig{
		Protocol:         "tcp",
		Address:          []byte("0.0.0.0:4043"),
		ExternalIp:       "0.0.0.0:4043",
		Peers:            []string{"0.0.0.0:30303"},
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       1024 * 4,
		WireHeaderLength: 9,
		TimeoutInMs:      200,
	}
	m := newP2PModule()

	{
		err := m.initialize(config)
		assert.Nil(
			t,
			err,
			"Error: failed to initialize p2p peer. %s", err,
		)
	}

	{
		go m.listen()
		_, waiting := <-m.ready

		assert.Equalf(
			t,
			waiting,
			false,
			"Request error: gater still not started after Listen",
		)
		t.Log("Pong: p2p peer started listening...")
	}

	addr := "127.0.0.1:22312"
	ready, _, data, _ := ListenAndServe(addr, int(config.BufferSize), 200)

	{
		select {
		case v := <-ready:
			assert.NotEqual(
				t,
				v,
				0,
				"Send error: could not start recipient server",
			)
		}
		t.Log("Pong: mock peer started lisetning...")
	}

	pongnonce := uint32(1)
	{
		err := m.pong(pongnonce, []byte(types.PING), addr)
		assert.Nil(
			t,
			err,
			"Pong: failed to send a pong message. Error: %s", err,
		)
		t.Log("Pong: p2p peer has pong'd the simulated ping sequence")
		t.Logf("Pong: p2p peer has ponged successfully.")
	}

	{
		t.Logf("Pong: mock peer waiting on pong sequence...")
		d := <-data
		t.Log("Pong: mock peer received the pong sequence")

		nonce, _, buff, _, err := (&wireCodec{}).decode(d.buff)
		assert.Nilf(
			t,
			err,
			"Pong: mock peer failed to decode wire bytes received from pong sequence. Error: %s", err,
		)

		pong := string(buff)
		assert.Equalf(
			t,
			pong,
			types.PONG,
			"Pong: mock peer expected to receive a pong sequence: %s, got: %s instead.", types.PONG, pong,
		)

		assert.Equalf(
			t,
			nonce,
			pongnonce,
			"Pong: mock peer received wrong nonce, expected %d, got: %d", pongnonce, nonce,
		)
	}
}
