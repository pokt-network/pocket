package pre_persistence

import (
	"bytes"
	"testing"

	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func TestGetAllParams(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	expected := typesGenesis.DefaultParams()
	err := ctx.(*PrePersistenceContext).SetParams(expected)
	require.NoError(t, err)
	params, err := ctx.(*PrePersistenceContext).GetParams(0)
	require.NoError(t, err)
	height, err := ctx.(*PrePersistenceContext).GetHeight()
	require.NoError(t, err)
	fee, err := ctx.GetStringParam(types.MessagePauseServiceNodeFee, height)
	require.NoError(t, err)
	if params.BlocksPerSession != expected.BlocksPerSession ||
		fee != expected.MessagePauseServiceNodeFee ||
		!bytes.Equal(params.MessageChangeParameterFeeOwner, params.MessageChangeParameterFeeOwner) {
		t.Fatalf("wrong params, expected %v got %v", expected, params)
	}
}
