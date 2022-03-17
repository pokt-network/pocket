package types

type Request struct {
	Nonce       uint32
	ResponsesCh chan Packet
}

func (r *Request) Respond(packet Packet) {
	r.ResponsesCh <- packet
}
