package pre_persistence

import (
	"bytes"
	"testing"
)

func TestGetAllParams(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	expected := DefaultParams()
	err := ctx.(*PrePersistenceContext).SetParams(expected)
	if err != nil {
		t.Fatal(err)
	}
	params, err := ctx.(*PrePersistenceContext).GetParams(0)
	if err != nil {
		t.Fatal(err)
	}
	fee, err := ctx.GetMessagePauseServiceNodeFee()
	if err != nil {
		t.Fatal(err)
	}
	if params.BlocksPerSession != expected.BlocksPerSession ||
		fee != expected.MessagePauseServiceNodeFee ||
		!bytes.Equal(params.MessageChangeParameterFeeOwner, params.MessageChangeParameterFeeOwner) {
		t.Fatalf("wrong params, expected %v got %v", expected, params)
	}
}
