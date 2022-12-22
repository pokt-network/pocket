package p2p

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/p2p/providers"
	persABP "github.com/pokt-network/pocket/p2p/providers/addrbook_provider/persistence"
	"github.com/pokt-network/pocket/p2p/raintree"
	"github.com/pokt-network/pocket/p2p/stdnetwork"
	"github.com/pokt-network/pocket/p2p/transport"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/telemetry"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.P2PModule = &p2pModule{}

const (
	p2pModuleName = "p2p"
)

type p2pModule struct {
	bus    modules.Bus
	p2pCfg modules.P2PConfig // TODO (olshansky): to remove this since it'll be available via the bus

	listener typesP2P.Transport
	address  cryptoPocket.Address

	network typesP2P.Network

	injectedAddrBookProvider      providers.AddrBookProvider
	injectedCurrentHeightProvider providers.CurrentHeightProvider
}

func Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	return new(p2pModule).Create(runtimeMgr)
}

// IMPROVE: need to define a better pattern for dependency injection. Currently we are probably limiting ourselves by having a common constructor `Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error)` for all modules.
func CreateWithProviders(runtimeMgr modules.RuntimeMgr, addrBookProvider providers.AddrBookProvider, currentHeightProvider providers.CurrentHeightProvider) (modules.Module, error) {
	log.Println("Creating network module")
	var m *p2pModule

	cfg := runtimeMgr.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	p2pCfg := cfg.GetP2PConfig()

	privateKey, err := cryptoPocket.NewPrivateKey(p2pCfg.GetPrivateKey())
	if err != nil {
		return nil, err
	}
	m = &p2pModule{
		p2pCfg:                        p2pCfg,
		address:                       privateKey.Address(),
		injectedAddrBookProvider:      addrBookProvider,
		injectedCurrentHeightProvider: currentHeightProvider,
	}

	if !p2pCfg.GetIsClientOnly() {
		l, err := transport.CreateListener(p2pCfg)
		if err != nil {
			return nil, err
		}
		m.listener = l
	}

	return m, nil
}

func (*p2pModule) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	log.Println("Creating network module")
	var m *p2pModule

	cfg := runtimeMgr.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	p2pCfg := cfg.GetP2PConfig()

	privateKey, err := cryptoPocket.NewPrivateKey(p2pCfg.GetPrivateKey())
	if err != nil {
		return nil, err
	}
	m = &p2pModule{
		p2pCfg:  p2pCfg,
		address: privateKey.Address(),
	}

	if !p2pCfg.GetIsClientOnly() {
		l, err := transport.CreateListener(p2pCfg)
		if err != nil {
			return nil, err
		}
		m.listener = l
	}

	return m, nil
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

func (m *p2pModule) GetModuleName() string {
	return p2pModuleName
}

func (m *p2pModule) Start() error {
	log.Println("Starting network module")

	addrbookProvider := getAddrBookProvider(m)
	currentHeightProvider := getCurrentHeightProvider(m)

	if m.p2pCfg.GetUseRainTree() {
		m.network = raintree.NewRainTreeNetwork(m.address, m.GetBus(), m.p2pCfg, addrbookProvider, currentHeightProvider)
	} else {
		m.network = stdnetwork.NewNetwork(m.GetBus(), m.p2pCfg, addrbookProvider, currentHeightProvider)
	}

	if m.p2pCfg.GetIsClientOnly() {
		return nil
	}

	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterRegister(
			telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_NAME,
			telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_DESCRIPTION,
		)

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
		CounterIncrement(telemetry.P2P_NODE_STARTED_TIMESERIES_METRIC_NAME)

	return nil
}

func getAddrBookProvider(m *p2pModule) providers.AddrBookProvider {
	var addrbookProvider providers.AddrBookProvider
	if m.injectedAddrBookProvider == nil {
		addrbookProvider = persABP.NewPersistenceAddrBookProvider(m.GetBus(), m.p2pCfg)
	} else {
		addrbookProvider = m.injectedAddrBookProvider
	}
	return addrbookProvider
}

func getCurrentHeightProvider(m *p2pModule) providers.CurrentHeightProvider {
	var currentHeightProvider providers.CurrentHeightProvider
	if m.injectedCurrentHeightProvider == nil {
		currentHeightProvider = m.GetBus().GetConsensusModule()
	} else {
		currentHeightProvider = m.injectedCurrentHeightProvider
	}
	return currentHeightProvider
}

func (m *p2pModule) Stop() error {
	log.Println("Stopping network module")
	if err := m.listener.Close(); err != nil {
		return err
	}
	return nil
}

func (m *p2pModule) Broadcast(msg *anypb.Any) error {
	c := &messaging.PocketEnvelope{
		Content: msg,
	}
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(c)
	if err != nil {
		return err
	}
	log.Println("broadcasting message to network")

	return m.network.NetworkBroadcast(data)
}

func (m *p2pModule) Send(addr cryptoPocket.Address, msg *anypb.Any) error {
	c := &messaging.PocketEnvelope{
		Content: msg,
	}
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(c)
	if err != nil {
		return err
	}

	return m.network.NetworkSend(data, addr)
}

// TECHDEBT(drewsky): Discuss how to best expose/access `Address` throughout the codebase.
func (m *p2pModule) GetAddress() (cryptoPocket.Address, error) {
	return m.address, nil
}

func (*p2pModule) ValidateConfig(cfg modules.Config) error {
	// TODO (#334): implement this
	return nil
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

	networkMessage := messaging.PocketEnvelope{}
	if err := proto.Unmarshal(appMsgData, &networkMessage); err != nil {
		log.Println("Error decoding network message: ", err)
		return
	}

	event := messaging.PocketEnvelope{
		Content: networkMessage.Content,
	}

	m.GetBus().PublishEventToBus(&event)
}
