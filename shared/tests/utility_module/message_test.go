package utility_module

import (
	"github.com/pokt-network/pocket/utility/types"
	"testing"
)

func NewTestingSendMessage(_ *testing.T, fromAddress, toAddress []byte, amount string) types.MessageSend {
	return types.MessageSend{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
	}
}
