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
	modules.TelemetryModule
	telemetryOn bool

	bus    modules.Bus
	config *config.P2PConfig
	node   P2PNode
}

var _ modules.P2PModule = &p2pModule{}

func Create(config *config.Config) (modules.P2PModule, error) {
	cfg := map[string]interface{}{
		"id":               config.P2P.ID,
		"address":          config.P2P.ExternalIp,
		"readBufferSize":   int(config.P2P.BufferSize),
		"writeBufferSize":  int(config.P2P.BufferSize),
		"redundancy":       config.P2P.Redundancy,
		"peers":            config.P2P.Peers,
		"enable_telemetry": config.P2P.EnableTelemetry,
	}
	m := &p2pModule{
		config:      config.P2P,
		bus:         nil,
		node:        CreateP2PNode(cfg),
		telemetryOn: config.P2P.EnableTelemetry,
	}

	return m, nil
}

func (m *p2pModule) Start() error {
	m.node.Info("Starting p2p module...")

	m.RegisterCounterMetric(
		"nodes",
		"the counter to track the number of nodes online",
	)

	m.RegisterCounterMetric(
		"p2p_msg_received_total",
		"the counter to track received messages",
	)

	m.RegisterCounterMetric(
		"p2p_connections_opened_total",
		"the counter to track how many connections were open",
	)

	m.RegisterCounterMetric(
		"p2p_connections_closed_total",
		"the counter to track how many connections were open",
	)

	m.RegisterCounterMetric(
		"p2p_connections_pooled_total",
		"the counter to track how many connections were pooled",
	)

	m.RegisterCounterMetric(
		"p2p_msg_broadcast_failed_total",
		"the counter to track how many broadcast rounds were initiated",
	)

	m.RegisterCounterMetric(
		"p2p_msg_broadcast_succeeded_total",
		"the counter to track how many broadcast rounds were performed successfully",
	)

	m.RegisterCounterMetric(
		"p2p_msg_send_failed_total",
		"the counter to track how many message sends have failed",
	)

	m.RegisterCounterMetric(
		"p2p_msg_send_succeeded_total",
		"the counter to track how many message sends have succeeded",
	)

	m.RegisterCounterMetric(
		"p2p_msg_handle_failed_total",
		"the counter to track how many message handlings have failed",
	)

	m.RegisterCounterMetric(
		"p2p_msg_handle_skipped_total",
		"the counter to track how many messages handlings have succeeded",
	)

	m.RegisterCounterMetric(
		"p2p_msg_handle_succeeded_total",
		"the counter to track how many messages handlings have succeeded",
	)

	m.RegisterCounterMetric(
		"p2p_msg_broadcast_depth",
		"the counter to track the depths to which the broadcast algorithm has went",
	)

	telemetry, err := m.GetTelemetry()
	if err == nil && telemetry != nil {
		m.node.SetTelemetry(telemetry)
	}

	if m.bus != nil {
		m.node.OnNewMessage(func(msg *types.P2PMessage) {
			m.node.Info("Publishing")
			m.bus.PublishEventToBus(msg.Payload)
			m.IncrementCounterMetric("p2p_msg_handle_succeeded_total")
		})
	} else {
		m.node.Warn("PocketBus is not initialized; no events will be published")
	}

	err = m.node.Start()

	if err != nil {
		return err
	}

	go m.node.Handle()

	m.IncrementCounterMetric("nodes")

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
	if err := m.node.BroadcastMessage(msg, true, 0); err != nil {
		return err
	}

	return nil
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
	if err := m.node.WriteMessage(0, tcpAddr, msg); err != nil {
		m.IncrementCounterMetric("p2p_msg_send_failed_total")
		return err
	}

	m.IncrementCounterMetric("p2p_msg_send_succeeded_total")
	return nil
}

func (m *p2pModule) GetTelemetry() (modules.TelemetryModule, error) {
	m.node.Debug("telemetry enabled=%v", m.telemetryOn)
	if !m.telemetryOn {
		return nil, nil
	}

	bus := m.GetBus()
	if bus == nil {
		return nil, fmt.Errorf("PocketBus is not initialized")
	}

	telemetry := bus.GetTelemetryModule()

	return telemetry, nil
}

func (m *p2pModule) RegisterCounterMetric(name string, description string) {
	if m.telemetryOn {
		if telemetry, err := m.GetTelemetry(); err == nil && telemetry != nil {
			telemetry.RegisterCounterMetric(name, description)
		}
	}
}

func (m *p2pModule) IncrementCounterMetric(name string) {
	if m.telemetryOn {
		if telemetry, err := m.GetTelemetry(); err == nil && telemetry != nil {
			telemetry.IncrementCounterMetric(name)
		}
	} else {
		m.node.Info("Telemetry is OFF")
	}
}
