package ibc

import (
	"fmt"
	"sync"

	ibcTypes "github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
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

	cfg *configs.IBCConfig
	m   sync.Mutex

	logger *modules.Logger

	// Only a single host is allowed at a time
	host *host
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(ibcModule).Create(bus, options...)
}

func (m *ibcModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	*m = ibcModule{
		cfg:    bus.GetRuntimeMgr().GetConfig().IBC,
		logger: logger.Global.CreateLoggerForModule(modules.IBCModuleName),
	}
	m.logger.Info().Msg("🪐 creating IBC module 🪐")

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	// Only validators can be an IBC host due to the need for reliability
	isValidator := false
	if _, err := m.GetBus().GetUtilityModule().GetValidatorModule(); err == nil {
		isValidator = true
	}
	if isValidator && m.cfg.Enabled {
		m.logger.Info().Msg("🛰️ creating IBC host 🛰️")
		if err := m.newHost(); err != nil {
			m.logger.Error().Err(err).Msg("❌ failed to create IBC host")
			return nil, err
		}
	}

	return m, nil
}

func (m *ibcModule) Start() error {
	if !m.cfg.Enabled {
		m.logger.Info().Msg("🚫 IBC module disabled 🚫")
		return nil
	}
	m.logger.Info().Msg("🪐 starting IBC module 🪐")
	// TODO: start the host logic
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

	case messaging.IbcMessageContentType:
		msg, err := codec.GetCodec().FromAny(message)
		if err != nil {
			return err
		}
		ibcMessage, ok := msg.(*ibcTypes.IbcMessage)
		if !ok {
			return fmt.Errorf("failed to cast message to IBCMessage")
		}
		return m.handleIBCMessage(ibcMessage)

	default:
		return coreTypes.ErrUnknownIBCMessageType(string(message.MessageName()))
	}
}

// handleIBCMessage unpacks the IBC message to its type and calls the appropriate handler
func (m *ibcModule) handleIBCMessage(message *ibcTypes.IbcMessage) error {
	switch msg := message.Event.(type) {
	case *ibcTypes.IbcMessage_Update:
		return m.handleUpdateMessage(msg.Update)
	case *ibcTypes.IbcMessage_Prune:
		return m.handlePruneMessage(msg.Prune)
	default:
		return coreTypes.ErrUnknownIBCMessageType(fmt.Sprintf("%T", msg))
	}
}

// handleUpdateMessage adds the updated store entry to the IBC store change mempool
func (m *ibcModule) handleUpdateMessage(message *ibcTypes.UpdateIbcStore) error {
	if m.host == nil {
		return coreTypes.ErrHostDoesNotExist()
	}
	// TODO: implement this
	return nil
}

// handlePruneMessage adds a removal entry to the IBC store change mempool
func (m *ibcModule) handlePruneMessage(message *ibcTypes.PruneIbcStore) error {
	if m.host == nil {
		return coreTypes.ErrHostDoesNotExist()
	}
	// TODO: implement this
	return nil
}

// newHost returns a new IBC host instance if one is not already created
func (m *ibcModule) newHost() error {
	if m.host != nil {
		return coreTypes.ErrHostAlreadyExists()
	}
	host := &host{
		logger: m.logger,
	}
	m.host = host
	return nil
}
