package utility

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/shared/utils"
	"github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

func TestUtilityContext_HandleMessageSend(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	accs := getAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := utils.BigIntToString(sendAmount)
	senderBalanceBefore, err := utils.StringToBigInt(accs[0].GetAmount())
	require.NoError(t, err)

	recipientBalanceBefore, err := utils.StringToBigInt(accs[1].GetAmount())
	require.NoError(t, err)

	addrBz, er := hex.DecodeString(accs[0].GetAddress())
	require.NoError(t, er)

	addrBz2, er := hex.DecodeString(accs[1].GetAddress())
	require.NoError(t, er)

	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	err = ctx.handleMessageSend(&msg)
	require.NoError(t, err, "handle message send")

	accs = getAllTestingAccounts(t, ctx)
	senderBalanceAfter, err := utils.StringToBigInt(accs[0].GetAmount())
	require.NoError(t, err)

	recipientBalanceAfter, err := utils.StringToBigInt(accs[1].GetAmount())
	require.NoError(t, err)
	require.Equal(t, sendAmount, big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter))
	require.Equal(t, sendAmount, big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore))
}

func TestUtilityContext_GetMessageSendSignerCandidates(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	accs := getAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := utils.BigIntToString(sendAmount)

	addrBz, er := hex.DecodeString(accs[0].GetAddress())
	require.NoError(t, er)

	addrBz2, er := hex.DecodeString(accs[1].GetAddress())
	require.NoError(t, er)

	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	candidates, err := ctx.getMessageSendSignerCandidates(&msg)
	require.NoError(t, err)
	require.Equal(t, 1, len(candidates))
	require.Equal(t, addrBz, candidates[0])
}

func NewTestingSendMessage(_ *testing.T, fromAddress, toAddress []byte, amount string) types.MessageSend {
	return types.MessageSend{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
	}
}
