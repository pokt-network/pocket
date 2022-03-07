package types

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
)

type Peer struct {
	id      uint64
	address string
}

func NewPeer(id uint64, addr string) *Peer {
	return &Peer{
		id:      id,
		address: addr,
	}
}

func (p *Peer) GenerateId() error {
	idbytes := make([]byte, 8)

	_, err := rand.Read(idbytes[:])

	if err != nil {
		return err
	}

	id, n := binary.Uvarint(idbytes)

	if n == 0 {
		return errors.New("Peer error: cannot generate id, buffer too small")
	}

	if n < 0 {
		return errors.New("Peer error: cannot generate id, buffer overflow")
	}

	p.id = id

	return nil
}

func (p *Peer) Id() uint64   { return p.id }
func (p *Peer) Addr() string { return p.address }
