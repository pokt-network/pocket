package types

import (
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestVoteValidateBasic(t *testing.T) {
	publicKey, err := crypto.GeneratePublicKey()
	require.NoError(t, err)
	testHash := crypto.SHA3Hash([]byte("fake_hash"))
	v := createLegacyVote(publicKey.Bytes(), 1, DoubleSignEvidenceType, testHash)
	require.NoError(t, v.ValidateBasic())
	// bad public key
	v2 := createLegacyVote([]byte("not_a_public_key"), 1, DoubleSignEvidenceType, testHash)
	badPkLen := len(v2.PublicKey)
	require.Equal(t, v2.ValidateBasic(), ErrInvalidPublicKeyLen(crypto.ErrInvalidPublicKeyLen(badPkLen)))
	// no public key
	v2.PublicKey = nil
	require.Equal(t, v2.ValidateBasic(), ErrEmptyPublicKey())
	// bad hash
	v3 := createLegacyVote(publicKey.Bytes(), 1, DoubleSignEvidenceType, []byte("not_a_hash"))
	badBlockHashLen := len(v3.BlockHash)
	require.Equal(t, v3.ValidateBasic(), ErrInvalidHashLength(crypto.ErrInvalidHashLen(badBlockHashLen)))
	// no hash
	v3.BlockHash = nil
	require.Equal(t, v3.ValidateBasic(), ErrEmptyHash())
	// negative height
	v4 := createLegacyVote(publicKey.Bytes(), -1, DoubleSignEvidenceType, testHash)
	v4.Height = -1
	require.Equal(t, v4.ValidateBasic(), ErrInvalidBlockHeight())
	// bad type
	v5 := createLegacyVote(publicKey.Bytes(), 0, 1, testHash)
	v5.Type = 0
	require.Equal(t, v5.ValidateBasic(), ErrInvalidEvidenceType())
}

func createLegacyVote(pubKey []byte, height int64, typ uint32, hash []byte) *LegacyVote {
	return &LegacyVote{
		PublicKey: pubKey,
		Height:    height,
		Round:     2,
		Type:      typ,
		BlockHash: hash,
	}
}
