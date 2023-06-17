package ibc

import (
	"fmt"
	"sync"

	"github.com/pokt-network/pocket/ibc/stores"
	"github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.IBCModule = &ibcModule{}

type ibcModule struct {
	base_modules.IntegratableModule

	m      sync.Mutex
	logger *modules.Logger

	// If the IBC module is enabled AND the node is a validator then a host will be created
	// otherwise this module will be disabled
	enabled bool

	// Only a single host is allowed at a time
	host *host
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(ibcModule).Create(bus, options...)
}

func (m *ibcModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m.logger = logger.Global.CreateLoggerForModule(m.GetModuleName())

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()

	ibcCfg := runtimeMgr.GetConfig().IBC
	m.enabled = false
	if runtimeMgr.GetConfig().Validator.Enabled && ibcCfg.Enabled {
		m.enabled = true
	}

	return m, nil
}

func (m *ibcModule) Start() error {
	if !m.enabled {
		return nil
	}
	m.logger.Info().Msg("🪐 starting IBC module 🪐")
	m.logger.Info().Msg("🛰️ creating IBC host 🛰️")
	_, err := m.newHost()
	if err != nil {
		m.logger.Error().Err(err).Msg("❌ failed to create IBC host")
		return err
	}
	return nil
}

func (m *ibcModule) Stop() error {
	return nil
}

func (m *ibcModule) GetHost() modules.IBCHost {
	return m.host
}

func (m *ibcModule) GetModuleName() string {
	return modules.IBCModuleName
}

// HandleMessage accepts a generic IBC message and routes it to the specific handler
func (m *ibcModule) HandleMessage(message *anypb.Any) error {
	m.m.Lock()
	defer m.m.Unlock()

	switch message.MessageName() {

	case messaging.IBCMessageContentType:
		msg, err := codec.GetCodec().FromAny(message)
		if err != nil {
			return err
		}
		ibcMessage, ok := msg.(*types.IBCMessage)
		if !ok {
			return fmt.Errorf("failed to cast message to IBCMessage")
		}
		return m.handleIBCMessage(ibcMessage)

	default:
		return coreTypes.ErrUnknownIBCMessageType(string(message.MessageName()))
	}
}

// handleIBCMessage unpacks the IBC message to its type and calls the appropriate handler
func (m *ibcModule) handleIBCMessage(message *types.IBCMessage) error {
	switch msg := message.Msg.(type) {
	case *types.IBCMessage_Update:
		return m.handleUpdateMessage(msg.Update)
	case *types.IBCMessage_Prune:
		return m.handlePruneMessage(msg.Prune)
	default:
		return coreTypes.ErrUnknownIBCMessageType(fmt.Sprintf("%T", msg))
	}
}

// handleUpdateMessage updates the appropriate IBC prefixed store with the given key/value pair
func (m *ibcModule) handleUpdateMessage(message *types.UpdateIBCStore) error {
	if m.host == nil {
		return coreTypes.ErrHostDoesNotExist()
	}
	store, err := m.host.GetStoreManager().GetStore(string(message.Prefix))
	if err != nil {
		return err
	}
	return store.Set(message.Key, message.Value)
}

// handlePruneMessage deletes the given key from the appropriate IBC prefixed store
func (m *ibcModule) handlePruneMessage(message *types.PruneIBCStore) error {
	if m.host == nil {
		return coreTypes.ErrHostDoesNotExist()
	}
	store, err := m.host.GetStoreManager().GetStore(string(message.Prefix))
	if err != nil {
		return err
	}
	return store.Delete(message.Key)
}

// newHost returns a new IBC host instance if one is not already created
func (m *ibcModule) newHost() (modules.IBCHost, error) {
	if m.host != nil {
		return nil, coreTypes.ErrHostAlreadyExists()
	}

	host := &host{
		logger: m.logger,
		stores: stores.NewStoreManager(m.storesDir),
	}

	m.host = host

	return host, nil
}
