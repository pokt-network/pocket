package p2p

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	persABP "github.com/pokt-network/pocket/p2p/providers/addrbook_provider/persistence"
	rpcABP "github.com/pokt-network/pocket/p2p/providers/addrbook_provider/rpc"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	rpcCHP "github.com/pokt-network/pocket/p2p/providers/current_height_provider/rpc"
	"github.com/pokt-network/pocket/p2p/raintree"
	"github.com/pokt-network/pocket/p2p/stdnetwork"
	"github.com/pokt-network/pocket/p2p/transport"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/telemetry"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	modules.BaseIntegratableModule

	listener typesP2P.Transport
	address  cryptoPocket.Address

	logger modules.Logger

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
	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()
	p2pCfg := cfg.P2P

	configureBootstrapNodes(p2pCfg, m)

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

func (m *p2pModule) HandleEvent(event *anypb.Any) error {
	evt, err := codec.GetCodec().FromAny(event)
	if err != nil {
		return err
	}

	switch event.MessageName() {
	case messaging.ConsensusNewHeightEventType:
		consensusNewHeightEvent, ok := evt.(*messaging.ConsensusNewHeightEvent)
		if !ok {
			return fmt.Errorf("failed to cast event to ConsensusNewHeightEvent")
		}

		addrBook := m.network.GetAddrBook()
		newAddrBook, err := m.addrBookProvider.GetStakedAddrBookAtHeight(consensusNewHeightEvent.Height)

		if err != nil {
			return err
		}

		added, removed := getAddrBookDelta(addrBook, newAddrBook)
		for _, add := range added {
			if err := m.network.AddPeerToAddrBook(add); err != nil {
				return err
			}
		}
		for _, rm := range removed {
			if err := m.network.RemovePeerToAddrBook(rm); err != nil {
				return err
			}
		}

	case messaging.StateMachineTransitionEventType:
		stateMachineTransitionEvent, ok := evt.(*messaging.StateMachineTransitionEvent)
		if !ok {
			return fmt.Errorf("failed to cast event to StateMachineTransitionEvent")
		}

		if stateMachineTransitionEvent.Dst == "P2P_bootstrapping" {
			addrBook := m.network.GetAddrBook()
			if len(addrBook) == 0 {
				m.logger.Warn().Msg("No peers in addrbook, bootstrapping") // TODO: deblasis - fix this. For some reason it's not logging
				log.Println("No peers in addrbook, bootstrapping")

				err := bootstrap(m)
				if err != nil {
					return err
				}
			}
			if !isSelfInAddrBook(m.address, addrBook) {
				m.logger.Warn().Msg("Self address not found in addresbook, advertising") // TODO: deblasis - fix this. For some reason it's not logging
				log.Println("Self address not found in addresbook, advertising")
				// TODO: advertise node to network, populate internal addressbook adding self as first peer
			}
			if err := m.GetBus().GetStateMachineModule().Event(context.TODO(), "P2P_isBootstrapped"); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unknown event type: %s", event.MessageName())
	}

	return nil

}

func configureBootstrapNodes(p2pCfg *configs.P2PConfig, m *p2pModule) {
	if p2pCfg.BootstrapNodesCsv == "" {
		m.bootstrapNodes = strings.Split(defaults.DefaultP2PBootstrapNodesCsv, ",")
	} else {
		m.bootstrapNodes = strings.Split(p2pCfg.BootstrapNodesCsv, ",")
	}
}

func isSelfInAddrBook(selfAddr cryptoPocket.Address, addrBook typesP2P.AddrBook) bool {
	for _, peer := range addrBook {
		if peer.Address.Equals(selfAddr) {
			return true
		}
	}
	return false
}

func bootstrap(m *p2pModule) error {
	var (
		addrBook typesP2P.AddrBook
	)

	for _, bootstrapNode := range m.bootstrapNodes {
		m.logger.Info().Str("endpoint", bootstrapNode).Msg("Attempting to bootstrap from bootstrap node") // TODO: deblasis - fix this. For some reason it's not logging
		log.Println("Attempting to bootstrap from bootstrap node: " + bootstrapNode)

		client, err := rpc.NewClientWithResponses(bootstrapNode)
		if err != nil {
			continue
		}
		healthCheck, err := client.GetV1Health(context.TODO())
		if err != nil || healthCheck == nil || healthCheck.StatusCode != http.StatusOK {
			log.Println("Error getting a green health check from bootstrap node: " + bootstrapNode)
			continue
		}

		addressBookProvider := rpcABP.NewRPCAddrBookProvider(
			rpcABP.WithP2PConfig(
				m.GetBus().GetRuntimeMgr().GetConfig().P2P,
			),
			rpcABP.WithCustomRPCUrl(bootstrapNode),
		)

		currentHeightProvider := rpcCHP.NewRPCCurrentHeightProvider(rpcCHP.WithCustomRPCUrl(bootstrapNode))

		addrBook, err = addressBookProvider.GetStakedAddrBookAtHeight(currentHeightProvider.CurrentHeight())
		if err != nil {
			m.logger.Warn().Err(err).Str("endpoint", bootstrapNode).Msg("Error getting address book from bootstrap node") // TODO: deblasis - fix this. For some reason it's not logging
			log.Println("Error getting address book from bootstrap node: " + bootstrapNode)
			continue
		}
	}

	if len(addrBook) == 0 {
		return fmt.Errorf("bootstrap failed")
	}

	for _, peer := range addrBook {
		log.Println("Adding peer to addrBook: " + peer.Address.String())
		if err := m.network.AddPeerToAddrBook(peer); err != nil {
			return err
		}
	}
	return nil
}
