package types

import (
	"bytes"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestUtilityCodec(t *testing.T) {
	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)
	v := UnstakingActor{
		Address:       addr,
		StakeAmount:   "100",
		OutputAddress: addr,
	}
	v2 := UnstakingActor{}
	codec := GetCodec()
	protoBytes, err := codec.Marshal(&v)
	require.NoError(t, err)
	if err := codec.Unmarshal(protoBytes, &v2); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v.Address, v2.Address) || v.StakeAmount != v2.StakeAmount || !bytes.Equal(v.OutputAddress, v2.OutputAddress) {
		t.Fatalf("unequal objects after marshal/unmarshal, expected %v, got %v", v, v2)
	}
	any, err := codec.ToAny(&v)
	require.NoError(t, err)
	protoMsg, err := codec.FromAny(any)
	require.NoError(t, err)
	v3, ok := protoMsg.(*UnstakingActor)
	if !ok {
		t.Fatal("any couldn't be converted back to original type")
	}
	if !bytes.Equal(v.Address, v3.Address) || v.StakeAmount != v3.StakeAmount || !bytes.Equal(v.OutputAddress, v3.OutputAddress) {
		t.Fatalf("unequal objects after any/fromAny, expected %v, got %v", v, v2)
	}
}
