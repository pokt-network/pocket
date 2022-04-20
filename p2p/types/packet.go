package types

type Packet struct {
	Nonce     uint32
	Data      []byte
	From      string // TODO(team): Change this temporary string when we strar using IDs
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
