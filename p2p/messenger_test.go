package p2p

import (
	"pocket/p2p/types"
	"testing"
)

func TestMessager_Protobuff(t *testing.T) {
	nonce, src, dst, level := int32(1), "10.0.0.1:1234", "9.0.0.2:5432", int32(2)

	// test message instantiation
	{
		p2pmsg := Message(nonce, level, types.PocketTopic_P2P_PING, src, dst)

		if p2pmsg.Topic != types.PocketTopic_P2P_PING {
			t.Errorf("Protobuff messenger error: Failed to instantiate a ping message, expected topic: %s, got: %s", types.PocketTopic_P2P_PING, p2pmsg.Topic)
		}

		if p2pmsg.Source != src {
			t.Errorf("Protobuff messenger error: Failed to instantiate a ping message, expected source: %s, got: %s", src, p2pmsg.Source)
		}

		if p2pmsg.Destination != dst {
			t.Errorf("Protobuff messenger error: Failed to instantiate a ping message, expected destination: %s, got: %s", dst, p2pmsg.Destination)
		}

		if p2pmsg.Nonce != nonce {
			t.Errorf("Protobuff messenger error: Failed to instantiate a ping message, expected nonce: %d, got: %d", nonce, p2pmsg.Nonce)
		}

		if p2pmsg.Level != level {
			t.Errorf("Protobuff messenger error: Failed to instantiate a ping message, expected level: %d, got: %d", level, p2pmsg.Level)
		}

	}

	// test encoding/decoding of ping
	{
		p2pmsg := Message(nonce, level, types.PocketTopic_P2P_PING, src, dst)

		encoded, err := Encode(*p2pmsg)

		if err != nil {
			t.Errorf("Protobuff messenger error: failed to encode ping message: %s", err.Error())
		}

		if len(encoded) == 0 {
			t.Errorf("Protobuff messenger error: corrupted encoding, encryption length is 0.")
		}

		msg, err := Decode(encoded)
		p2pmsg = &msg

		if err != nil {
			t.Errorf("Protobuff messenger error: failed to decode ping message: %s", err.Error())
		}

		if p2pmsg.Nonce != nonce {
			t.Errorf("Protobuff messenger error: decoder corrupted ping message, expected nonce: %d, got: %d", nonce, p2pmsg.Nonce)
		}

		if p2pmsg.Level != level {
			t.Errorf("Protobuff messenger error: decoder corrupted ping message, expected level: %d, got: %d", level, p2pmsg.Level)
		}

		if p2pmsg.Source != src {
			t.Errorf("Protobuff messenger error: decoder corrupted ping message, expected source: %s, got: %s", src, p2pmsg.Source)
		}

		if p2pmsg.Destination != dst {
			t.Errorf("Protobuff messenger error: decoder corrupted ping message, expected action: %s, got: %s", dst, p2pmsg.Destination)
		}

		if p2pmsg.Topic != types.PocketTopic_P2P_PING {
			t.Errorf("Protobuff messenger error: decoder corrupted ping message, expected topic: %s, got: %s", types.PocketTopic_P2P_PING, p2pmsg.Topic)
		}
	}
}
