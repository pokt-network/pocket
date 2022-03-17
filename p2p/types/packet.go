package types

type Packet struct {
	Nonce     uint32
	Data      []byte
	From      string // temporary until we start using ids
	isEncoded bool   // should decode data using the domain codec or not
}

func NewPacket(nonce uint32, data []byte, from string, isEncoded bool) Packet {
	return Packet{
		Nonce:     nonce,
		Data:      data,
		From:      from,
		isEncoded: isEncoded,
	}
}
