package p2p

import (
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

type p2pModule struct {
	bus modules.Bus

	listener typesP2P.Transport
	address  cryptoPocket.Address

	network typesP2P.Network

	injectedAddrBookProvider      providers.AddrBookProvider
	injectedCurrentHeightProvider providers.CurrentHeightProvider
}

func Create(bus modules.Bus) (modules.Module, error) {
	return new(p2pModule).Create(bus)
}

// TODO(#429): need to define a better pattern for dependency injection. Currently we are probably limiting ourselves by having a common constructor `Create(bus modules.Bus) (modules.Module, error)` for all modules.
func CreateWithProviders(bus modules.Bus, addrBookProvider providers.AddrBookProvider, currentHeightProvider providers.CurrentHeightProvider) (modules.Module, error) {
	log.Println("Creating network module")
	m := &p2pModule{}
	if err := bus.RegisterModule(m); err != nil {
		return nil, err
	}

	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()
	p2pCfg := cfg.P2P

	privateKey, err := cryptoPocket.NewPrivateKey(p2pCfg.GetPrivateKey())
	if err != nil {
		return nil, err
	}
	m.address = privateKey.Address()
	m.injectedAddrBookProvider = addrBookProvider
	m.injectedCurrentHeightProvider = currentHeightProvider

	if !cfg.ClientDebugMode {
		l, err := transport.CreateListener(p2pCfg)
		if err != nil {
			return nil, err
		}
		m.listener = l
	}

	return m, nil
}

func (*p2pModule) Create(bus modules.Bus) (modules.Module, error) {
	log.Println("Creating network module")
	m := &p2pModule{}
	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()
	p2pCfg := cfg.P2P

	privateKey, err := cryptoPocket.NewPrivateKey(p2pCfg.PrivateKey)
	if err != nil {
		return nil, err
	}
	m.address = privateKey.Address()

	if !cfg.ClientDebugMode {
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
	return modules.P2PModuleName
}

func (m *p2pModule) Start() error {
	log.Println("Starting network module")

	addrbookProvider := getAddrBookProvider(m)
	currentHeightProvider := getCurrentHeightProvider(m)

	cfg := m.GetBus().GetRuntimeMgr().GetConfig()

	if cfg.P2P.UseRainTree {
		m.network = raintree.NewRainTreeNetwork(m.address, m.GetBus(), addrbookProvider, currentHeightProvider)
	} else {
		m.network = stdnetwork.NewNetwork(m.GetBus(), addrbookProvider, currentHeightProvider)
	}

	if cfg.ClientDebugMode {
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

// CLEANUP(#429): marked for removal since we'll implement a better pattern for dependency injection
func getAddrBookProvider(m *p2pModule) providers.AddrBookProvider {
	var addrbookProvider providers.AddrBookProvider
	if m.injectedAddrBookProvider == nil {
		addrbookProvider = persABP.NewPersistenceAddrBookProvider(m.GetBus())
	} else {
		addrbookProvider = m.injectedAddrBookProvider
	}
	return addrbookProvider
}

// CLEANUP(#429): marked for removal since we'll implement a better pattern for dependency injection
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

func (m *p2pModule) handleNetworkMessage(networkMsgData []byte) {
	appMsgData, err := m.network.HandleNetworkData(networkMsgData)
	if err != nil {
		log.Println("Error handling raw data: ", err)
		return
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if appMsgData == nil {
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
