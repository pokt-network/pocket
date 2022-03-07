package p2p

import (
	"encoding/binary"
	"testing"
)

func TestTypesCodec_Encode(t *testing.T) {
	c := NewTypesCodec()

	msg := struct {
		topic  string
		action string
	}{
		topic:  "membership",
		action: "ping",
	}

	encoder := func(m struct {
		topic  string
		action string
	}) ([]byte, error) {
		topic := []byte(m.topic)
		action := []byte(m.action)

		tlen := uint16(len(topic))
		alen := uint16(len(action))

		tlenb := make([]byte, 2)
		binary.BigEndian.PutUint16(tlenb[:2], tlen)

		alenb := make([]byte, 2)
		binary.BigEndian.PutUint16(alenb[:2], alen)

		buff := make([]byte, 0)

		buff = append(buff, tlenb...)
		buff = append(buff, alenb...)
		buff = append(buff, topic...)
		buff = append(buff, action...)

		return buff, nil
	}

	decoder := func(d []byte) (struct {
		topic  string
		action string
	}, error) {
		tlenb := d[:2]

		tlen := binary.BigEndian.Uint16(tlenb)

		msg := d[4:]
		topic := string(msg[:tlen])
		action := string(msg[tlen:])

		return struct {
			topic  string
			action string
		}{topic: topic, action: action}, nil
	}

	c.Register(msg, encoder, decoder)

	encoding, err := c.Encode(msg)
	if err != nil {
		t.Errorf("Domain Codec error: failed to encode message: %s", err.Error())
	}

	expectedlen := 2 + 2 + len(msg.topic) + len(msg.action)
	if len(encoding) == expectedlen {
		t.Errorf("Domain Codec error: wrong encoding byte length, expected %d, got %d", expectedlen, len(encoding))
	}

	decoding, err := decoder(encoding[2:])
	if err != nil {
		t.Errorf("Domain Codec error: failed to decode encoding bytes")
	}

	m := decoding

	if m.topic != msg.topic && m.action != msg.action {
		t.Errorf("Domain Codec error: corrupted decoded message, expected %s, %s, got %s, %s", msg.topic, msg.action, m.topic, m.action)
	}
}

func TestTypesCodec_Decode(t *testing.T) {
	c := NewTypesCodec()

	msg := struct {
		topic  string
		action string
	}{
		topic:  "membership",
		action: "ping",
	}

	encoder := func(m struct {
		topic  string
		action string
	}) ([]byte, error) {
		topic := []byte(m.topic)
		action := []byte(m.action)

		tlen := uint16(len(topic))
		alen := uint16(len(action))

		tlenb := make([]byte, 2)
		binary.BigEndian.PutUint16(tlenb[:2], tlen)

		alenb := make([]byte, 2)
		binary.BigEndian.PutUint16(alenb[:2], alen)

		buff := make([]byte, 0)

		buff = append(buff, tlenb...)
		buff = append(buff, alenb...)
		buff = append(buff, topic...)
		buff = append(buff, action...)

		return buff, nil
	}

	decoder := func(d []byte) (struct {
		topic  string
		action string
	}, error) {
		tlenb := d[:2]

		tlen := binary.BigEndian.Uint16(tlenb)

		msg := d[4:]
		topic := string(msg[:tlen])
		action := string(msg[tlen:])

		return struct {
			topic  string
			action string
		}{topic: topic, action: action}, nil
	}

	id, err := c.Register(msg, encoder, decoder)
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb[:2], id)
	encoding, err := encoder(msg)
	decoding, err := c.Decode(append(idb, encoding...))
	if err != nil {
		t.Errorf("Domain Codec error: failed to decode encoding bytes: %s", err.Error())
	}

	m := decoding.(struct {
		topic  string
		action string
	})

	if m.topic != msg.topic && m.action != msg.action {
		t.Errorf("Domain Codec error: corrupted decoded message, expected %s, %s, got %s, %s", msg.topic, msg.action, m.topic, m.action)
	}
}
