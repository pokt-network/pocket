package pre2p

// TODO(team): This is a a temporary parallel to the real `p2p` module.
// It should be removed once the real `p2p` module is ready but is meant
// to be a "real" replacement for now.

import (
	"github.com/pokt-network/pocket/p2p/pre2p/raintree"
	"github.com/pokt-network/pocket/p2p/pre2p/stdnetwork"
	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"

	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/logging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.P2PModule = &p2pModule{}

type p2pModule struct {
	bus modules.Bus

	listener typesPre2P.Transport
	address  cryptoPocket.Address

	network typesPre2P.Network

	// used to keep around this configuration value until `Start` is called.
	// That's when a logger is registered for this module, and that's when we need to properly set its level
	logLevel logging.LogLevel
}

func Create(cfg *config.Config) (m modules.P2PModule, err error) {
	logging.GetGlobalLogger().Info("pre2p: Creating network module")

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

	pre2pLogLevel := logging.GetLevel(cfg.Pre2P.LogLevel)

	if pre2pLogLevel == logging.LOG_LEVEL_NONE {
		logging.GetGlobalLogger().Warn("pre2p: Configuration error: Invalid log level supplied. Suppressing logs using level=NONE")
	}

	m = &p2pModule{
		listener: l,
		network:  network,
		address:  cfg.PrivateKey.Address(),
		logLevel: pre2pLogLevel,
	}

	return m, nil
}

func (m *p2pModule) SetBus(bus modules.Bus) {
	m.bus = bus
}

func (m *p2pModule) GetBus() modules.Bus {
	if m.bus == nil {
		logging.GetGlobalLogger().Warn("pre2p: PocketBus is not initialized")
		return nil
	}
	return m.bus
}

func (m *p2pModule) GetLogger() logging.Logger {
	if m.bus != nil {
		m.GetBus().GetTelemetryModule().LoggerGet(logging.P2P_NAMESPACE)
	}

	logging.GetGlobalLogger().Fatal("pre2p: Cannot retrieve pre2p bus logger. PocketBus is not initialized")
	return nil
}

func (m *p2pModule) Start() error {
	// DISCUSSION(team): how about exposing the logger, telemetry, and other modules functionality through injected-context.
	// I will reduce code footprint and redundancy significantly (ctx.logger.) (ctx.metrics.) (etc...)
	m.
		GetBus().
		GetTelemetryModule().
		LoggerRegister(logging.P2P_NAMESPACE, m.logLevel)

	m.GetLogger().Info("Starting network module")

	m.
		GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterRegister(
			"p2p_nodes_online",
			"the counter to track the number of nodes online",
		)

	m.network.SetBus(m.GetBus())

	go func() {
		for {
			data, err := m.listener.Read()
			if err != nil {
				m.GetLogger().Error("Error reading data from connection: ", err)
				continue
			}
			go m.handleNetworkMessage(data)
		}
	}()

	m.
		GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterIncrement("p2p_nodes_online")

	return nil
}

func (m *p2pModule) Stop() error {
	m.GetLogger().Info("Stopping network module")
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

	m.GetLogger().Info("broadcasting message to network")

	return m.network.NetworkBroadcast(data)
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

	return m.network.NetworkSend(data, addr)
}

func (m *p2pModule) handleNetworkMessage(networkMsgData []byte) {
	appMsgData, err := m.network.HandleNetworkData(networkMsgData)
	if err != nil {
		m.GetLogger().Error("Error handling raw data: ", err)
		return
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if appMsgData == nil {
		m.GetLogger().Debug("No app-specific message to forward from the network")
		return
	}

	networkMessage := types.PocketEvent{}
	if err := proto.Unmarshal(appMsgData, &networkMessage); err != nil {
		m.GetLogger().Error("Error decoding network message: ", err)
		return
	}

	event := types.PocketEvent{
		Topic: networkMessage.Topic,
		Data:  networkMessage.Data,
	}

	m.GetBus().PublishEventToBus(&event)
}
