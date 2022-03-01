package types

func (m *NetworkMessage) IsRequest() bool { return m.Nonce != 0 }
