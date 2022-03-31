package types

import (
	"bytes"
	"testing"
)

func TestUtilityCodec(t *testing.T) {
	v := Vote{
		PublicKey: []byte("pubkey"),
		Height:    1,
		Round:     2,
		Type:      3,
		BlockHash: []byte("hash"),
	}
	v2 := Vote{}
	codec := UtilityCodec()
	protoBytes, err := codec.Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}
	if err := codec.Unmarshal(protoBytes, &v2); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v.PublicKey, v2.PublicKey) || v.Height != v2.Height || v.Round != v2.Round || v.Type != v2.Type ||
		!bytes.Equal(v.BlockHash, v2.BlockHash) {
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
	v3, ok := protoMsg.(*Vote)
	if !ok {
		t.Fatal("any couldn't be converted back to original type")
	}
	if !bytes.Equal(v.PublicKey, v3.PublicKey) || v.Height != v3.Height || v.Round != v3.Round || v.Type != v3.Type ||
		!bytes.Equal(v.BlockHash, v3.BlockHash) {
		t.Fatalf("unequal objects after any/fromAny, expected %v, got %v", v, v2)
	}
}
