package test

import (
	"testing"

	"github.com/pokt-network/pocket/internal/utility/types"
)

func NewTestingSendMessage(_ *testing.T, fromAddress, toAddress []byte, amount string) types.MessageSend {
	return types.MessageSend{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
	}
}
