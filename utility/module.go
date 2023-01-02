package utility

import (
	"fmt"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	_ modules.UtilityModule = &utilityModule{}
	_ modules.UtilityConfig = &types.UtilityConfig{}
	_ modules.Module        = &utilityModule{}
)

type utilityModule struct {
	bus    modules.Bus
	config modules.UtilityConfig

	logger modules.Logger

	Mempool types.Mempool
}

const (
	utilityModuleName = "utility"

	TransactionGossipMessageContentType = "utility.TransactionGossipMessage"
)

func Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	return new(utilityModule).Create(runtime)
}

func (*utilityModule) Create(runtime modules.RuntimeMgr) (modules.Module, error) {
	var m *utilityModule

	cfg := runtime.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	utilityCfg := cfg.GetUtilityConfig()

	return &utilityModule{
		config:  utilityCfg,
		logger:  logger.Global.CreateLoggerForModule(m.GetModuleName()),
		Mempool: types.NewMempool(utilityCfg.GetMaxMempoolTransactionBytes(), utilityCfg.GetMaxMempoolTransactions()),
	}, nil
}

func (u *utilityModule) Start() error {
	return nil
}

func (u *utilityModule) Stop() error {
	return nil
}

func (u *utilityModule) GetModuleName() string {
	return utilityModuleName
}

func (u *utilityModule) SetBus(bus modules.Bus) {
	u.bus = bus
}

func (u *utilityModule) GetBus() modules.Bus {
	if u.bus == nil {
		u.logger.Fatal().Msg("Bus is not initialized")
	}
	return u.bus
}

func (*utilityModule) ValidateConfig(cfg modules.Config) error {
	// TODO (#334): implement this
	return nil
}

func (u *utilityModule) HandleMessage(message *anypb.Any) error {
	switch message.MessageName() {
	case TransactionGossipMessageContentType:
		msg, err := codec.GetCodec().FromAny(message)
		if err != nil {
			return err
		}
		transactionGossipMsg, ok := msg.(*types.TransactionGossipMessage)
		if !ok {
			return fmt.Errorf("failed to cast message to UtilityMessage")
		}
		if err := u.CheckTransaction(transactionGossipMsg.Tx); err != nil {
			return err
		}
		u.logger.Info().Str("source", "MEMPOOL").Msg("Successfully added a new message to the mempool!")
	default:
		return types.ErrUnknownMessageType(message.MessageName())
	}
	return nil
}
