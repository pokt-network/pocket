package p2p

import (
	"fmt"
	"log"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"

	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	shared "github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/types/known/anypb"

	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
)

type p2pModule struct {
	bus    modules.Bus
	config *config.P2PConfig
	node   P2PNode
}

var _ modules.P2PModule = &p2pModule{}

func Create(config *config.Config) (modules.P2PModule, error) {
	cfg := map[string]interface{}{
		"id":              config.P2P.ID,
		"address":         config.P2P.ExternalIp,
		"readBufferSize":  int(config.P2P.BufferSize),
		"writeBufferSize": int(config.P2P.BufferSize),
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
	var tcpAddr string
	v, exists := typesGenesis.GetNodeState(nil).ValidatorMap[addr.String()]
	if !exists {
		return fmt.Errorf("[ERROR]: p2p send: address not in validator map")
	}
	tcpAddr = v.ServiceUrl

	m.node.Info("Sending to:", tcpAddr)

	msg := types.NewP2PMessage(0, 0, m.node.Address(), tcpAddr, &shared.PocketEvent{
		Topic: topic,
		Data:  data,
	})
	return m.node.WriteMessage(0, tcpAddr, msg)
}
