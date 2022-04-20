package types

import (
	"bytes"
	"github.com/pokt-network/pocket/shared/crypto"
	"testing"
)

func TestUtilityCodec(t *testing.T) {
	addr, err := crypto.GenerateAddress()
	if err != nil {
		t.Fatal(err)
	}
	v := UnstakingActor{
		Address:       addr,
		StakeAmount:   "100",
		OutputAddress: addr,
	}
	v2 := UnstakingActor{}
	codec := GetCodec()
	protoBytes, err := codec.Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}
	if err := codec.Unmarshal(protoBytes, &v2); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v.Address, v2.Address) || v.StakeAmount != v2.StakeAmount || !bytes.Equal(v.OutputAddress, v2.OutputAddress) {
		t.Fatalf("unequal objects after marshal/unmarshal, expected %v, got %v", v, v2)
	}
	any, err := codec.ToAny(&v)
	if err != nil {
		t.Fatal(err)
	}
	protoMsg, err := codec.FromAny(any)
	if err != nil {
		t.Fatal(err)
	}
	v3, ok := protoMsg.(*UnstakingActor)
	if !ok {
		t.Fatal("any couldn't be converted back to original type")
	}
	if !bytes.Equal(v.Address, v3.Address) || v.StakeAmount != v3.StakeAmount || !bytes.Equal(v.OutputAddress, v3.OutputAddress) {
		t.Fatalf("unequal objects after any/fromAny, expected %v, got %v", v, v2)
	}
}
