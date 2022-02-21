package p2p

import (
	"pocket/p2p/types"

	"github.com/golang/protobuf/proto"
)

func Message(nonce int32, level int32, topic types.PocketTopic, src, dest string) *types.NetworkMessage {
	return &types.NetworkMessage{
		Level:       level,
		Nonce:       nonce,
		Topic:       topic,
		Source:      src,
		Destination: dest,
	}
}

func Encode(m types.NetworkMessage) ([]byte, error) {
	data, err := proto.Marshal(&m)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Decode(data []byte) (types.NetworkMessage, error) {
	msg := &types.NetworkMessage{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return types.NetworkMessage{Nonce: -1, Level: -1}, err
	}
	return *msg, nil
}
