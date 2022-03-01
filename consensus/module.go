package consensus

import (
	"log"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"

	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.ConsensusModule = &consensusModule{}

type consensusModule struct {
	bus modules.Bus
}

func Create(_ *config.Config) (modules.ConsensusModule, error) {
	m := &consensusModule{
		bus: nil,
	}
	return m, nil
}

func (m *consensusModule) Start() error {
	// TODO(olshansky): Add a test that bus is set
	log.Println("Starting consensus module...")
	return nil
}

func (m *consensusModule) Stop() error {
	log.Println("Stopping consensus module...")
	return nil
}

func (m *consensusModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *consensusModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *consensusModule) HandleMessage(anyMessage *anypb.Any) {
	panic("HandleMessage not implemented")
}
