package pre_p2p

import (
	"log"
	"pocket/shared/config"
	"pocket/shared/modules"

	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.NetworkModule = &networkModule{}

type networkModule struct {
	pocketBus modules.Bus
	p2pConfig *config.P2PConfig
}

func Create(config *config.Config) (modules.NetworkModule, error) {
	return &networkModule{
		pocketBus: nil,
		p2pConfig: config.P2P,
	}, nil
}

func (p *networkModule) Start() error {
	// TODO(olshansky): Add a test that pocketBus is set
	log.Println("Starting PRE P2P module...")
	return nil
}

func (p *networkModule) Stop() error {
	log.Println("Stopping PRE P2P module...")
	return nil
}

func (m *networkModule) SetBus(pocketBus modules.Bus) {
	m.pocketBus = pocketBus
}

func (m *networkModule) GetBus() modules.Bus {
	if m.pocketBus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBus
}

func (m *networkModule) BroadcastMessage(msg *anypb.Any, topic string) error {
	panic("BroadcastMessage not implemented")
}

func (m *networkModule) Send(addr string, msg *anypb.Any, topic string) error {
	panic("Send not implemented")
}
