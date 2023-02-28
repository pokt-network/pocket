package utility

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/utility/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/anypb"
)

func PrepareTxGossipMessage(txBz []byte) (*anypb.Any, error) {
	txGossipMessage := &typesUtil.TxGossipMessage{
		Tx: txBz,
	}

	// nolint:gocritic // TODO: keeping commented out code in place because this is how it should work in the future
	// pocketEnvelope, err := messaging.PackMessage(txGossipMessage)
	// if err != nil {
	// 	return nil, err
	// }

	anyMessage, err := codec.GetCodec().ToAny(txGossipMessage)
	if err != nil {
		return nil, err
	}

	return anyMessage, nil
}

func (u *utilityModule) HandleUtilityMessage(message *anypb.Any) error {
	switch message.MessageName() {
	case messaging.TxGossipMessageContentType:
		msg, err := codec.GetCodec().FromAny(message)
		if err != nil {
			return err
		}

		if txGossipMsg, ok := msg.(*types.TxGossipMessage); !ok {
			return fmt.Errorf("failed to cast message to UtilityMessage")
		} else if err := u.HandleTransaction(txGossipMsg.Tx); err != nil {
			return err
		}

		u.logger.Info().Str("message_type", "TxGossipMessage").Msg("Successfully added a new message to the mempool!")

	default:
		return types.ErrUnknownMessageType(message.MessageName())
	}

	return nil
}
