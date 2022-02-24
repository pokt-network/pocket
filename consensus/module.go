package consensus

import (
	"log"
	"pocket/shared/config"
	"pocket/shared/modules"

	"google.golang.org/protobuf/types/known/anypb"
)

const (
	DefaultLogPrefix string = "NODE" // Just a default that'll be replaced during consensus operations.
)

var _ modules.ConsensusModule = &consensusModule{}

type consensusModule struct {
	modules.ConsensusModule
	pocketBus modules.Bus
}

func Create(_ *config.Config) (modules.ConsensusModule, error) {
	m := &consensusModule{
		ConsensusModule: nil, // TODO(olshansky): sync with Andrew on a better way to do this
		pocketBus:       nil,
	}
	return m, nil
}

func (m *consensusModule) Start() error {
	// TODO(olshansky): Add a test that pocketBus is set
	log.Println("Starting consensus module...")
	return nil
}

func (m *consensusModule) Stop() error {
	log.Println("Stopping consensus module...")
	return nil
}

func (m *consensusModule) GetBus() modules.Bus {
	if m.pocketBus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBus
}

func (m *consensusModule) SetBus(pocketBus modules.Bus) {
	m.pocketBus = pocketBus
}

func (m *consensusModule) HandleMessage(anyMessage *anypb.Any) {
	panic("HandleMessage not implemented")
}

func (m *consensusModule) HandleTransaction(anyMessage *anypb.Any) {
	panic("HandleTransaction not implemented")
}
