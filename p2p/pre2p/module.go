package pre2p

// TODO(team): This is a a temporary parallel to the real `p2p` module.
// It should be removed once the real `p2p` module is ready but is meant
// to be a "real" replacement for now.

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/p2p/pre2p/raintree"
	"github.com/pokt-network/pocket/p2p/pre2p/stdnetwork"
	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"

	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	// p2pModule is a telemetryModule, that will use the shared telemetry module in the bus
	// this allows the p2pModule to have fine-grain control over its own telemetry, using the shared telemetry module.
	modules.TelemetryModule
	telemetryOn bool

	bus modules.Bus

	listener typesPre2P.TransportLayerConn
	address  cryptoPocket.Address

	network typesPre2P.Network
}

func Create(cfg *config.Config) (m modules.P2PModule, err error) {
	log.Println("Creating network module")

	l, err := CreateListener(cfg.Pre2P)
	if err != nil {
		return nil, err
	}

	testState := typesGenesis.GetNodeState(nil)
	addrBook, err := ValidatorMapToAddrBook(cfg.Pre2P, testState.ValidatorMap)
	if err != nil {
		return nil, err
	}

	var network typesPre2P.Network
	if cfg.Pre2P.UseRainTree {
		selfAddr := cfg.PrivateKey.Address()
		network = raintree.NewRainTreeNetwork(selfAddr, addrBook, cfg)
	} else {
		network = stdnetwork.NewNetwork(addrBook)
	}

	m = &p2pModule{
		listener:    l,
		network:     network,
		address:     cfg.PrivateKey.Address(),
		telemetryOn: cfg.Pre2P.EnableTelemetry,
	}

	return m, nil
}

func (m *p2pModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *p2pModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Printf("[WARN]: PocketBus is not initialized")
		return nil
	}
	return m.bus
}

func (m *p2pModule) Start() error {
	log.Println("Starting network module")

	m.RegisterCounterMetric(
		"nodes",
		"the counter to track the number of nodes online",
	)

	m.RegisterCounterMetric(
		"p2p_msg_received_total",
		"the counter to track received messages",
	)

	m.RegisterCounterMetric(
		"p2p_opened_connections_total",
		"the counter to track how many connections were open",
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
		"p2p_msg_handle_succeeded_total",
		"the counter to track how many messages handlings have succeeded",
	)

	m.RegisterCounterMetric(
		"p2p_msg_broadcast_depth",
		"the counter to track the depths to which the broadcast algorithm has went",
	)

	telemetry, err := m.GetTelemetry()
	if err == nil && telemetry != nil {
		m.network.SetTelemetry(telemetry)
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("Encoutered error", err)
			}
		}()
		for {
			data, err := m.listener.Read()
			if err != nil {
				log.Println("Error reading data from connection: ", err)
				continue
			}

			log.Printf("Was it this one?")
			m.IncrementCounterMetric("p2p_opened_connections_total")

			log.Printf("or that one?")
			m.IncrementCounterMetric("p2p_msg_received_total")

			go m.handleNetworkMessage(data)
		}
	}()

	// TODO(tema):
	// No way to know if the node has actually properly started as of the current implementation of m.listener
	// Make sure to do this after the node has started if the implementation allowed for this in the future.
	m.GetBus().
		GetTelemetryModule().
		IncrementCounterMetric("nodes")

	return nil
}

func (m *p2pModule) Stop() error {
	log.Println("Stopping network module")
	if err := m.listener.Close(); err != nil {
		return err
	}
	return nil
}

func (m *p2pModule) Broadcast(msg *anypb.Any, topic types.PocketTopic) error {
	c := &types.PocketEvent{
		Topic: topic,
		Data:  msg,
	}
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}
	log.Println("broadcasting message to network")

	if err := m.network.NetworkBroadcast(data); err != nil {
		m.IncrementCounterMetric("p2p_msg_broadcast_failed_total")
		return err
	}

	m.GetBus().
		GetTelemetryModule().
		IncrementCounterMetric("p2p_msg_broadcast_succeeded_total")

	return nil
}

func (m *p2pModule) Send(addr cryptoPocket.Address, msg *anypb.Any, topic types.PocketTopic) error {
	c := &types.PocketEvent{
		Topic: topic,
		Data:  msg,
	}
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}

	if err := m.network.NetworkSend(data, addr); err != nil {
		m.IncrementCounterMetric("p2p_msg_send_failed_total")
		return err
	}

	m.IncrementCounterMetric("p2p_msg_send_succeeded_total")

	return nil
}

func (m *p2pModule) handleNetworkMessage(networkMsgData []byte) {
	appMsgData, err := m.network.HandleNetworkData(networkMsgData)
	if err != nil {
		m.IncrementCounterMetric("p2p_msg_handle_failed_total")

		log.Println("Error handling raw data: ", err)

		return
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if appMsgData == nil {
		return
	}

	networkMessage := types.PocketEvent{}
	if err := proto.Unmarshal(appMsgData, &networkMessage); err != nil {
		m.IncrementCounterMetric("p2p_msg_handle_failed_total")

		log.Println("Error decoding network message: ", err)

		return
	}

	event := types.PocketEvent{
		Topic: networkMessage.Topic,
		Data:  networkMessage.Data,
	}

	m.GetBus().PublishEventToBus(&event)

	m.IncrementCounterMetric("p2p_msg_handle_succeeded_total")
}

func (m *p2pModule) GetTelemetry() (modules.TelemetryModule, error) {
	if m.telemetryOn {
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
	}
}
