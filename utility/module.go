package utility

import (
	"fmt"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/mempool"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	_ modules.UtilityModule = &utilityModule{}
	_ modules.Module        = &utilityModule{}
)

type utilityModule struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule

	config *configs.UtilityConfig

	logger  modules.Logger
	mempool mempool.TXMempool
}

const (
	TransactionGossipMessageContentType = "utility.TransactionGossipMessage"
)

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(utilityModule).Create(bus, options...)
}

func (*utilityModule) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &utilityModule{}

	for _, option := range options {
		option(m)
	}

	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()

	cfg := runtimeMgr.GetConfig()
	utilityCfg := cfg.Utility

	m.config = utilityCfg
	m.mempool = types.NewTxFIFOMempool(utilityCfg.MaxMempoolTransactionBytes, utilityCfg.MaxMempoolTransactions)

	return m, nil
}

func (u *utilityModule) Start() error {
	u.logger = logger.Global.CreateLoggerForModule(u.GetModuleName())
	return nil
}

func (u *utilityModule) GetModuleName() string {
	return modules.UtilityModuleName
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

func (u *utilityModule) GetMempool() mempool.TXMempool {
	return u.mempool
}
