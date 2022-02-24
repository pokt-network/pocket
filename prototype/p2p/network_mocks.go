package p2p

import "pocket/p2p/types"

func MockBasicInstance() *P2PModule {
	return &P2PModule{
		inbound:  *NewIoMap(0),
		outbound: *NewIoMap(0),

		c: NewDomainCodec(),

		peerlist: types.NewPeerlist(),
		sink:     make(chan types.Work, 1),

		done:    make(chan uint, 1),
		ready:   make(chan uint, 1),
		closed:  make(chan uint, 1),
		errored: make(chan uint, 1),
	}
}
