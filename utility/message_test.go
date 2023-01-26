package utility

import (
	"testing"

	"github.com/pokt-network/pocket/utility/types"
)

func NewTestingSendMessage(_ *testing.T, fromAddress, toAddress []byte, amount string) types.MessageSend {
	return types.MessageSend{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
	}
}
