package p2p

// TODO(team): This is a a temporary parallel to the real `p2p` module.
// It should be removed once the real `p2p` module is ready but is meant
// to be a "real" replacement for now.

import (
	"encoding/json"
	"github.com/pokt-network/pocket/shared/debug"
	"log"

	"github.com/pokt-network/pocket/p2p/raintree"
	"github.com/pokt-network/pocket/p2p/stdnetwork"
	p2pTelemetry "github.com/pokt-network/pocket/p2p/telemetry"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	bus       modules.Bus
	p2pConfig modules.P2PConfig // TODO (olshansky): to remove this since it'll be available via the bus

	listener typesP2P.Transport
	address  cryptoPocket.Address

	network typesP2P.Network
}

func (m *p2pModule) GetAddress() (cryptoPocket.Address, error) {
	return m.address, nil
}

func Create(config, gen json.RawMessage) (m modules.P2PModule, err error) {
	log.Println("Creating network module")
	cfg, err := InitConfig(config)
	if err != nil {
		return nil, err
	}
	l, err := CreateListener(cfg)
	if err != nil {
		return nil, err
	}
	privateKey, err := cryptoPocket.NewPrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}
	m = &p2pModule{
		p2pConfig: cfg,

		listener: l,
		address:  privateKey.Address(),

		network: nil,
	}

	return m, nil
}

func InitGenesis(data json.RawMessage) {
	// TODO (Team) add genesis if necessary
	return
}

func InitConfig(data json.RawMessage) (config *typesP2P.P2PConfig, err error) {
	config = new(typesP2P.P2PConfig)
	err = json.Unmarshal(data, config)
	return
}

func (m *p2pModule) SetBus(bus modules.Bus) {
	// INVESTIGATE: Can the code flow be modified to set the bus here?
	// m.network.SetBus(m.GetBus())
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

	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterRegister(
			p2pTelemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_NAME,
			p2pTelemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_DESCRIPTION,
		)

	addrBook, err := ValidatorMapToAddrBook(m.p2pConfig, m.bus.GetConsensusModule().ValidatorMap())
	if err != nil {
		return err
	}

	if m.p2pConfig.GetUseRainTree() {
		m.network = raintree.NewRainTreeNetwork(m.address, addrBook)
	} else {
		m.network = stdnetwork.NewNetwork(addrBook)
	}
	m.network.SetBus(m.GetBus())
	go func() {
		for {
			data, err := m.listener.Read()
			if err != nil {
				log.Println("Error reading data from connection: ", err)
				continue
			}
			go m.handleNetworkMessage(data)
		}
	}()

	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterIncrement(p2pTelemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_NAME)

	return nil
}

func (m *p2pModule) Stop() error {
	log.Println("Stopping network module")
	if err := m.listener.Close(); err != nil {
		return err
	}
	return nil
}

func (m *p2pModule) Broadcast(msg *anypb.Any, topic debug.PocketTopic) error {
	c := &debug.PocketEvent{
		Topic: topic,
		Data:  msg,
	}
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}
	log.Println("broadcasting message to network")

	return m.network.NetworkBroadcast(data)
}

func (m *p2pModule) Send(addr cryptoPocket.Address, msg *anypb.Any, topic debug.PocketTopic) error {
	c := &debug.PocketEvent{
		Topic: topic,
		Data:  msg,
	}
	data, err := proto.Marshal(c)
	if err != nil {
		return err
	}

	return m.network.NetworkSend(data, addr)
}

func (m *p2pModule) handleNetworkMessage(networkMsgData []byte) {
	appMsgData, err := m.network.HandleNetworkData(networkMsgData)
	if err != nil {
		log.Println("Error handling raw data: ", err)
		return
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if appMsgData == nil {
		// log.Println("[DEBUG] No app-specific message to forward from the network")
		return
	}

	networkMessage := debug.PocketEvent{}
	if err := proto.Unmarshal(appMsgData, &networkMessage); err != nil {
		log.Println("Error decoding network message: ", err)
		return
	}

	event := debug.PocketEvent{
		Topic: networkMessage.Topic,
		Data:  networkMessage.Data,
	}

	m.GetBus().PublishEventToBus(&event)
}
