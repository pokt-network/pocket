package types

import (
	"testing"

	"github.com/pokt-network/pocket/internal/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestVoteValidateBasic(t *testing.T) {
	publicKey, err := crypto.GeneratePublicKey()
	require.NoError(t, err)
	testHash := crypto.SHA3Hash([]byte("fake_hash"))
	v := LegacyVote{
		PublicKey: publicKey.Bytes(),
		Height:    1,
		Round:     2,
		Type:      DoubleSignEvidenceType,
		BlockHash: testHash,
	}
	require.NoError(t, v.ValidateBasic())
	// bad public key
	v2 := v
	v2.PublicKey = []byte("not_a_public_key")
	badPkLen := len(v2.PublicKey)
	require.Equal(t, v2.ValidateBasic(), ErrInvalidPublicKeyLen(crypto.ErrInvalidPublicKeyLen(badPkLen)))
	// no public key
	v2.PublicKey = nil
	require.Equal(t, v2.ValidateBasic(), ErrEmptyPublicKey())
	// bad hash
	v3 := v
	v3.BlockHash = []byte("not_a_hash")
	badBlockHashLen := len(v3.BlockHash)
	require.Equal(t, v3.ValidateBasic(), ErrInvalidHashLength(crypto.ErrInvalidHashLen(badBlockHashLen)))
	// no hash
	v3.BlockHash = nil
	require.Equal(t, v3.ValidateBasic(), ErrEmptyHash())
	// negative height
	v4 := v
	v4.Height = -1
	require.Equal(t, v4.ValidateBasic(), ErrInvalidBlockHeight())
	// bad type
	v5 := v
	v5.Type = 0
	require.Equal(t, v5.ValidateBasic(), ErrInvalidEvidenceType())
}
