package types

type Work struct {
	nonce   uint32
	data    []byte
	addr    string // temporary until we start using ids
	encoded bool   // should decode data using the domain codec or not
}

func (w *Work) Nonce() uint32 {
	return w.nonce
}

func (w *Work) isEncoded() bool {
	return w.encoded
}

func (w *Work) Bytes() []byte {
	return w.data
}

func (w *Work) From() string {
	return w.addr
}

func NewWork(nonce uint32, data []byte, addr string, encoded bool) Work {
	return Work{nonce, data, addr, encoded}
}

func (w *Work) Implode() (uint32, []byte, string, bool) {
	return w.nonce, w.data, w.addr, w.encoded
}
