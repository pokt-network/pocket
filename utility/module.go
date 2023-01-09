package utility

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ modules.UtilityModule = &utilityModule{}
var _ modules.Module = &utilityModule{}

type utilityModule struct {
	bus    modules.Bus
	config *configs.UtilityConfig

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
	utilityCfg := runtime.GetConfig().Utility

	return &utilityModule{
		config:  utilityCfg,
		Mempool: types.NewMempool(utilityCfg.MaxMempoolTransactionBytes, utilityCfg.MaxMempoolTransactions),
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
		log.Fatalf("Bus is not initialized")
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
		log.Println("MEMPOOOL: Successfully added a new message to the mempool!")
	default:
		return types.ErrUnknownMessageType(message.MessageName())
	}
	return nil
}
