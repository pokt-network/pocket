package p2p

import (
	"log"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	persABP "github.com/pokt-network/pocket/p2p/providers/addrbook_provider/persistence"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	"github.com/pokt-network/pocket/p2p/raintree"
	"github.com/pokt-network/pocket/p2p/stdnetwork"
	"github.com/pokt-network/pocket/p2p/transport"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"github.com/pokt-network/pocket/telemetry"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	base_modules.IntegratableModule

	listener typesP2P.Transport
	address  cryptoPocket.Address

	logger *modules.Logger

	network typesP2P.Network

	addrBookProvider      providers.AddrBookProvider
	currentHeightProvider providers.CurrentHeightProvider

	bootstrapNodes []string
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(p2pModule).Create(bus, options...)
}

func (*p2pModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	log.Println("Creating network module")
	m := &p2pModule{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()
	p2pCfg := cfg.P2P

	if err := m.configureBootstrapNodes(); err != nil {
		return nil, err
	}

	privateKey, err := cryptoPocket.NewPrivateKey(p2pCfg.PrivateKey)
	if err != nil {
		return nil, err
	}
	m.address = privateKey.Address()

	m.setupDependencies()

	if !cfg.ClientDebugMode {
		l, err := transport.CreateListener(p2pCfg)
		if err != nil {
			return nil, err
		}
		m.listener = l
	}

	return m, nil
}

func (m *p2pModule) setupDependencies() {
	addrBookProvider, err := m.GetBus().GetModulesRegistry().GetModule(addrbook_provider.ModuleName)
	if err != nil {
		addrBookProvider = persABP.NewPersistenceAddrBookProvider(m.GetBus())
	}
	m.addrBookProvider = addrBookProvider.(providers.AddrBookProvider)

	currentHeightProvider, err := m.GetBus().GetModulesRegistry().GetModule(current_height_provider.ModuleName)
	if err != nil {
		currentHeightProvider = m.GetBus().GetConsensusModule()
	}
	m.currentHeightProvider = currentHeightProvider.(providers.CurrentHeightProvider)
}

func (m *p2pModule) GetModuleName() string {
	return modules.P2PModuleName
}

func (m *p2pModule) Start() error {
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())
	m.logger.Info().Msg("Starting network module")

	cfg := m.GetBus().GetRuntimeMgr().GetConfig()

	// TODO: pass down logger
	if cfg.P2P.UseRainTree {
		m.network = raintree.NewRainTreeNetwork(m.address, m.GetBus(), m.addrBookProvider, m.currentHeightProvider)
	} else {
		m.network = stdnetwork.NewNetwork(m.GetBus(), m.addrBookProvider, m.currentHeightProvider)
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
				m.logger.Error().Err(err).Msg("Error reading data from connection")
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

func (m *p2pModule) Stop() error {
	m.logger.Info().Msg("Stopping network module")
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
	m.logger.Info().Msg("broadcasting message to network")

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
		m.logger.Error().Err(err).Msg("Error handling raw data")
		return
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if appMsgData == nil {
		return
	}

	networkMessage := messaging.PocketEnvelope{}
	if err := proto.Unmarshal(appMsgData, &networkMessage); err != nil {
		m.logger.Error().Err(err).Msg("Error decoding network message")
		return
	}

	event := messaging.PocketEnvelope{
		Content: networkMessage.Content,
	}

	m.GetBus().PublishEventToBus(&event)
}
