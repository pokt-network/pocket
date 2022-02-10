package p2p

import (
	"bytes"
	"encoding/gob"

	"pocket/consensus/pkg/p2p/p2p_types"
)

func EncodeNetworkMessage(message *p2p_types.NetworkMessage) ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(message); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func DecodeNetworkMessage(data []byte) (*p2p_types.NetworkMessage, error) {
	var buff = bytes.NewBuffer(data)
	dec := gob.NewDecoder(buff)
	networkMessage := &p2p_types.NetworkMessage{}
	if err := dec.Decode(networkMessage); err != nil {
		return nil, err
	}
	return networkMessage, nil
}
