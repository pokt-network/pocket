package utility

import (
	"fmt"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	_ modules.UtilityModule = &utilityModule{}
	_ modules.Module        = &utilityModule{}
)

type utilityModule struct {
	bus    modules.Bus
	config *configs.UtilityConfig

	logger modules.Logger

	Mempool types.Mempool
}

const (
	TransactionGossipMessageContentType = "utility.TransactionGossipMessage"
)

func Create(bus modules.Bus) (modules.Module, error) {
	return new(utilityModule).Create(bus)
}

func (*utilityModule) Create(bus modules.Bus) (modules.Module, error) {
	m := &utilityModule{}
	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()

	cfg := runtimeMgr.GetConfig()
	utilityCfg := cfg.Utility

	m.config = utilityCfg
	m.Mempool = types.NewMempool(utilityCfg.MaxMempoolTransactionBytes, utilityCfg.MaxMempoolTransactions)

	return m, nil
}

func (u *utilityModule) Start() error {
	u.logger = logger.Global.CreateLoggerForModule(u.GetModuleName())
	return nil
}

func (u *utilityModule) Stop() error {
	return nil
}

func (u *utilityModule) GetModuleName() string {
	return modules.UtilityModuleName
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
