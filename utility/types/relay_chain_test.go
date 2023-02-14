package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_RelayChain_Validate(t *testing.T) {
	relayChainValid := relayChain("0001")
	err := relayChainValid.ValidateBasic()
	require.NoError(t, err)

	relayChainInvalidLength := relayChain("001")
	expectedError := ErrInvalidRelayChainLength(0, relayChainLength)
	err = relayChainInvalidLength.ValidateBasic()
	require.Equal(t, expectedError.Code(), err.Code())

	relayChainEmpty := relayChain("")
	expectedError = ErrEmptyRelayChain()
	err = relayChainEmpty.ValidateBasic()
	require.Equal(t, expectedError.Code(), err.Code())
}
