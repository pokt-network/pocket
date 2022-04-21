package types

import shared "github.com/pokt-network/pocket/shared/types"

func (m *P2PMessage) Copy() *P2PMessage {
	newm := &P2PMessage{}

	newm.Metadata = &Metadata{
		Level:       m.Metadata.Level,
		Source:      m.Metadata.Source,
		Destination: m.Metadata.Destination,
		Broadcast:   m.Metadata.Broadcast,
	}

	newm.Payload = &shared.PocketEvent{
		Topic: m.Payload.Topic,
		Data:  m.Payload.Data,
	}

	return newm
}
