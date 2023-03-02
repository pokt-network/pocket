package p2p

import (
	"io"

	"github.com/pokt-network/pocket/shared/crypto"
)

type Peer interface {
	GetAddress() crypto.Address
	GetPublicKey() crypto.PublicKey
	GetServiceURL() string

	// TECHDEBT: move this to some new `ConnManager` interface.
	GetStream() io.ReadWriteCloser
}

type PeerList []Peer
