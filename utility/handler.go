package utility

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	TransactionGossipMessageContentType = "utility.TransactionGossipMessage"
)

func (u *utilityModule) HandleMessage(message *anypb.Any) error {
	switch message.MessageName() {
	case TransactionGossipMessageContentType:
		msg, err := codec.GetCodec().FromAny(message)
		if err != nil {
			return err
		}

		if transactionGossipMsg, ok := msg.(*types.TransactionGossipMessage); !ok {
			return fmt.Errorf("failed to cast message to UtilityMessage")
		} else if err := u.CheckTransaction(transactionGossipMsg.Tx); err != nil {
			return err
		}

		log.Println("MEMPOOL: Successfully added a new message to the mempool!")

	default:
		return types.ErrUnknownMessageType(message.MessageName())
	}

	return nil
}
