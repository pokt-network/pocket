package ibc

import (
	"fmt"

	"github.com/pokt-network/pocket/ibc/host"
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
	base_modules.IntegrableModule

	cfg    *configs.IBCConfig
	logger *modules.Logger

	// Only a single host is allowed at a time
	host modules.IBCHostSubmodule
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(ibcModule).Create(bus, options...)
}

func (m *ibcModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	*m = ibcModule{
		cfg:    bus.GetRuntimeMgr().GetConfig().IBC,
		logger: logger.Global.CreateLoggerForModule(modules.IBCModuleName),
	}
	m.logger.Info().Msg("🪐 Creating IBC module 🪐")

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
		if err := m.newHost(); err != nil {
			m.logger.Error().Err(err).Msg("❌ Failed to create IBC host ❌")
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
	m.logger.Info().Msg("✅ Starting IBC Module ✅")
	return nil
}

func (m *ibcModule) Stop() error {
	m.logger.Info().Msg("🛑 Stopping IBC Module 🛑")
	return nil
}

func (m *ibcModule) GetModuleName() string {
	return modules.IBCModuleName
}

func (m *ibcModule) HandleEvent(event *anypb.Any) error {
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
		newHeight := consensusNewHeightEvent.GetHeight()
		// check if host was created exit if not
		if m.host == nil {
			break
		}
		// Flush all caches to disk for last height
		bsc := m.GetBus().GetBulkStoreCacher()
		if err := bsc.FlushAllEntries(); err != nil {
			return err
		}
		// Prune old cache entries
		if newHeight <= m.cfg.Host.BulkStoreCacher.MaxHeightStored {
			break
		}
		pruneHeight := newHeight - m.cfg.Host.BulkStoreCacher.MaxHeightStored
		if err := bsc.PruneCaches(pruneHeight); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled event type: %s", event.MessageName())
	}
	return nil
}

// newHost creates a new IBC host and sets it in the ibcModule struct if it is not already set
func (m *ibcModule) newHost() error {
	if m.host != nil {
		return coreTypes.ErrIBCHostAlreadyExists()
	}
	hostMod, err := host.Create(m.GetBus(),
		m.cfg.Host,
		host.WithLogger(m.logger),
		host.WithStoresDir(m.cfg.StoresDir),
	)
	if err != nil {
		return err
	}
	m.host = hostMod
	return nil
}
