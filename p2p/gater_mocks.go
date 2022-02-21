package p2p

func MockGater() *gater {
	return &gater{
		inbound:  *NewIoMap(0),
		outbound: *NewIoMap(0),

		c: NewDomainCodec(),

		peerlist: &plist{elements: make([]peer, 0)},
		sink:     make(chan work, 1),

		listener: nil,

		err:    nil,
		done:   make(chan uint, 1),
		ready:  make(chan uint, 1),
		closed: make(chan uint, 1),
	}
}
