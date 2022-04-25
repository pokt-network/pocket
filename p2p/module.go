package p2p

import (
	"log"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"

	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	shared "github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/types/known/anypb"
)

type p2pModule struct {
	bus    modules.Bus
	config *config.P2PConfig
	node   P2PNode
}

var _ modules.P2PModule = &p2pModule{}

func Create(config *config.Config) (modules.P2PModule, error) {
	cfg := map[string]interface{}{
		"address":         config.P2P.ExternalIp,
		"readBufferSize":  int(config.P2P.BufferSize),
		"writeBufferSize": int(config.P2P.BufferSize),
		"id":              config.P2P.ID,
		"redundancy":      config.P2P.Redundancy,
		"peers":           config.P2P.Peers,
	}
	m := &p2pModule{
		config: config.P2P,
		bus:    nil,
		node:   CreateP2PNode(cfg),
	}

	return m, nil
}

func (m *p2pModule) Start() error {
	m.node.Info("Starting p2p module...")

	if m.bus != nil {
		m.node.OnNewMessage(func(msg *types.P2PMessage) {
			m.node.Info("Publishing")
			m.bus.PublishEventToBus(msg.Payload)
		})
	} else {
		m.node.Warn("PocketBus is not initialized; no events will be published")
	}

	err := m.node.Start()

	if err != nil {
		return err
	}

	go m.node.Handle()

	return nil
}

func (m *p2pModule) Stop() error {
	m.node.Stop()
	return nil
}

func (m *p2pModule) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *p2pModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *p2pModule) Broadcast(data *anypb.Any, topic shared.PocketTopic) error {
	msg := types.NewP2PMessage(0, 0, m.node.Address(), "", &shared.PocketEvent{
		Topic: topic,
		Data:  data,
	})

	msg.MarkAsBroadcastMessage()
	return m.node.BroadcastMessage(msg, true, 0)
}

func (m *p2pModule) Send(addr cryptoPocket.Address, data *anypb.Any, topic shared.PocketTopic) error {
	msg := types.NewP2PMessage(0, 0, m.node.Address(), string(addr), &shared.PocketEvent{
		Topic: topic,
		Data:  data,
	})
	return m.node.SendMessage(0, string(addr), msg)
}
