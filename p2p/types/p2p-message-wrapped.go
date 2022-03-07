package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/pokt-network/pocket/p2p/utils"
	shared "github.com/pokt-network/pocket/shared/types"
)

func Message(nonce int32, level int32, src, dest string, event *shared.PocketEvent) *P2PMessage {
	return &P2PMessage{
		Metadata: &Metadata{
			Hash:        utils.GenerateRandomHash(),
			Level:       level,
			Nonce:       nonce,
			Source:      src,
			Destination: dest,
		},
		Payload: event,
	}
}

func Encode(m P2PMessage) ([]byte, error) {
	data, err := proto.Marshal(&m)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Decode(data []byte) (P2PMessage, error) {
	msg := &P2PMessage{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return P2PMessage{}, err
	}
	return *msg, nil
}
