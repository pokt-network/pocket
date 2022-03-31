package p2p

import (
	"log"

	"github.com/pokt-network/pocket/shared/config"
	pcrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	bus       modules.Bus
	p2pConfig *config.P2PConfig
}

func Create(config *config.Config) (modules.P2PModule, error) {
	return &p2pModule{
		bus:       nil,
		p2pConfig: config.P2P,
	}, nil
}

func (p *p2pModule) Start() error {
	// TODO(olshansky): Add a test that bus is set
	log.Println("Starting P2P module...")
	return nil
}

func (p *p2pModule) Stop() error {
	log.Println("Stopping P2P module...")
	return nil
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
	panic("Broadcast not implemented")
}

func (m *p2pModule) Send(addr pcrypto.Address, msg *anypb.Any, topic types.PocketTopic) error {
	panic("Send not implemented")
}
