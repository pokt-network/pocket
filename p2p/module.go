package p2p

import (
	"log"

	"github.com/pokt-network/pocket/shared/config"
	pcrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/types/known/anypb"
)

type p2pModule struct {
	bus modules.Bus
}

var _ modules.P2PModule = &p2pModule{}

func Create(config *config.Config) (modules.P2PModule, error) {
	return &p2pModule{}, nil
}

func (m *p2pModule) Start() error {
	panic("Not implemented")
}

func (m *p2pModule) Stop() error {
	panic("Not implemented")
}

func (m *p2pModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *p2pModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *p2pModule) Broadcast(msg *anypb.Any, topic types.PocketTopic) error {
	panic("Not implemented")
}

func (m *p2pModule) Send(addr pcrypto.Address, data *anypb.Any, topic types.PocketTopic) error {
	panic("Not implemented")
}
