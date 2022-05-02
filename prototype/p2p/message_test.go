package p2p

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestChurnMgmtMessenger(t *testing.T) {
	cm := &churnmgmt{}
	nonce, src, dst, level := uint32(1), "10.0.0.1:1234", "9.0.0.2:5432", uint16(2)

	// test message instantiation
	{
		ping := cm.message(nonce, Ping, level, src, dst)
		pong := cm.message(nonce, Pong, level, dst, src)

		if ping.topic != Churn {
			t.Errorf("Churn management messenger error: Failed to instantiate a ping message, expected topic: %s, got: %s", Churn, ping.topic)
		}

		if ping.action != Ping {
			t.Errorf("Churn management messenger error: Failed to instantiate a ping message, expected action: %s, got: %s", Ping, ping.action)
		}

		if ping.source != src {
			t.Errorf("Churn management messenger error: Failed to instantiate a ping message, expected source: %s, got: %s", src, ping.source)
		}

		if ping.destination != dst {
			t.Errorf("Churn management messenger error: Failed to instantiate a ping message, expected destination: %s, got: %s", dst, ping.destination)
		}

		if ping.nonce != nonce {
			t.Errorf("Churn management messenger error: Failed to instantiate a ping message, expected nonce: %d, got: %d", nonce, ping.nonce)
		}

		if ping.level != level {
			t.Errorf("Churn management messenger error: Failed to instantiate a ping message, expected level: %d, got: %d", level, ping.level)
		}

		if pong.topic != Churn {
			t.Errorf("Churn management messenger error: Failed to instantiate a pong message, expected topic: %s, got: %s", Churn, ping.topic)
		}

		if pong.action != Pong {
			t.Errorf("Churn management messenger error: Failed to instantiate a pong message, expected action: %s, got: %s", Pong, ping.action)
		}

		if pong.source != dst {
			t.Errorf("Churn management messenger error: Failed to instantiate a ping message, expected source: %s, got: %s", dst, pong.source)
		}

		if pong.destination != src {
			t.Errorf("Churn management messenger error: Failed to instantiate a ping message, expected source: %s, got: %s", src, ping.destination)
		}

		if pong.nonce != nonce {
			t.Errorf("Churn management messenger error: Failed to instantiate a ping message, expected source: %d, got: %d", nonce, pong.nonce)
		}

	}

	// test encoding/decoding of ping
	{
		ping := cm.message(nonce, Ping, level, src, dst)

		encoded, err := cm.encode(ping)

		if err != nil {
			t.Errorf("Churn management messenger error: failed to encode ping message: %s", err.Error())
		}

		metadata := encoded[:12]
		header := encoded[12 : 12+8+2]
		body := encoded[12+8+2:]

		meta := make([]byte, 0)
		meta = append(meta, parseipstring(src)...)
		meta = append(meta, parseipstring(dst)...)
		if bytes.Compare(metadata, meta) != 0 {
			t.Errorf("Churn management Messenger error: corrupted encoding, corrupted source and destination ips in encoding metadata.")
		}

		nb := make([]byte, 4)
		binary.BigEndian.PutUint32(nb, uint32(nonce))
		if bytes.Compare(header[:4], nb) != 0 {
			t.Errorf("Churn management Messenger error: corrupted encoding, corrupted nonce in encoding header")
		}

		lb := make([]byte, 2)
		binary.BigEndian.PutUint16(lb, uint16(level))
		if bytes.Compare(header[4:6], lb) != 0 {
			t.Errorf("Churn management Messenger error: corrupted encoding, corrupted level in encoding header")
		}

		tl := make([]byte, 2)
		binary.BigEndian.PutUint16(tl, uint16(len([]byte(Churn))))
		if bytes.Compare(header[6:8], tl) != 0 {
			t.Errorf("Churn management Messenger error: corrupted encoding, corrupted topic len in encoding header")
		}

		al := make([]byte, 2)
		binary.BigEndian.PutUint16(al, uint16(len([]byte(Ping))))
		if bytes.Compare(header[8:10], al) != 0 {
			t.Errorf("Churn management Messenger error: corrupted encoding, corrupted topic len in encoding header")
		}

		if bytes.Compare(body[:len([]byte(Churn))], []byte(Churn)) != 0 {
			t.Errorf("Churn management messenger error: corrupted encoding, corrupted topic in encoding body")
		}

		if bytes.Compare(body[len([]byte(Churn)):], []byte(Ping)) != 0 {
			t.Errorf("Churn management messenger error: corrupted encoding, corrupted action in encoding body")
		}

		msg, err := cm.decode(encoded)

		if err != nil {
			t.Errorf("Churn management messenger error: failed to decode ping message: %s", err.Error())
		}

		if message(msg).nonce != nonce {
			t.Errorf("Churn management messenger error: decoder corrupted ping message, expected nonce: %d, got: %d", nonce, message(msg).nonce)
		}

		if message(msg).level != level {
			t.Errorf("Churn management messenger error: decoder corrupted ping message, expected level: %d, got: %d", level, message(msg).level)
		}

		if message(msg).source != src {
			t.Errorf("Churn management messenger error: decoder corrupted ping message, expected source: %s, got: %s", src, message(msg).source)
		}

		if message(msg).destination != dst {
			t.Errorf("Churn management messenger error: decoder corrupted ping message, expected action: %s, got: %s", dst, message(msg).destination)
		}

		if message(msg).topic != Churn {
			t.Errorf("Churn management messenger error: decoder corrupted ping message, expected topic: %s, got: %s", Churn, message(msg).topic)
		}

		if message(msg).action != Ping {
			t.Errorf("Churn management messenger error: decoder corrupted ping message, expected action: %s, got: %s", Ping, message(msg).action)
		}
	}
}
