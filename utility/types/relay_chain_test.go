package types

import (
	"github.com/pokt-network/pocket/shared/types"
	"testing"
)

func TestRelayChain_Validate(t *testing.T) {
	relayChainValid := RelayChain("0001")
	relayChainInvalidLength := RelayChain("001")
	relayChainEmpty := RelayChain("")
	if err := relayChainValid.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := relayChainInvalidLength.Validate(); err.Code() != types.ErrInvalidRelayChainLength(0, RelayChainLength).Code() {
		t.Fatal(err)
	}
	if err := relayChainEmpty.Validate(); err.Code() != types.ErrEmptyRelayChain().Code() {
		t.Fatal(err)
	}
}
