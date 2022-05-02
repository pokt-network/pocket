package vrf

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVRFKeygenWithoutSeed(t *testing.T) {
	sk, vk, err := GenerateVRFKeys(nil)
	require.Nil(t, err)
	require.NotNil(t, sk)
	require.NotNil(t, vk)
}

func TestVRFKeygenWithSeed(t *testing.T) {
	seed := "abcdefghijklmnopqrstuvwxyz123456"
	require.Equal(t, len(seed), ed25519.SeedSize)

	privKey := ed25519.NewKeyFromSeed([]byte(seed))
	lastBlockHash := seed

	reader, err := CreateVRFRandReader(lastBlockHash, &privKey)
	require.Nil(t, err)
	require.NotNil(t, reader)

	sk, vk, err := GenerateVRFKeys(reader)
	require.Nil(t, err)

	require.Equal(t, "7172737475767778797a3132333435360000000000000000000000000000000035e96c98e934872f1998cb7c62536d710708523b893f139fe6acd11d3f519e08", hex.EncodeToString(sk.Bytes()))
	require.Equal(t, "35e96c98e934872f1998cb7c62536d710708523b893f139fe6acd11d3f519e08", hex.EncodeToString(vk.Bytes()))
}

func TestVRFKeygenProveAndVerify(t *testing.T) {
	seed := "abcdefghijklmnopqrstuvwxyz123456"
	require.Equal(t, len(seed), ed25519.SeedSize)

	privKey := ed25519.NewKeyFromSeed([]byte(seed))
	lastBlockHash := seed

	reader, err := CreateVRFRandReader(lastBlockHash, &privKey)
	require.Nil(t, err)
	require.NotNil(t, reader)

	msg := []byte("Proving this is HotPocket")

	sk, vk, err := GenerateVRFKeys(reader)
	require.Nil(t, err)

	vrfOut, vrfProof, err := sk.Prove(msg)
	require.Nil(t, err)
	assert.Equal(t, "cc7c0cf2099c4261774bef72eea749a30ed8ecccda900166e66e1695526a2e67a7164fd35f67df656c4ff92db90076106466db4b57af30effe2e2c67d66d7603", hex.EncodeToString(vrfOut))
	assert.Equal(t, "064ca43b1eb7a0bd2dfac3c8cd74b44468d65a985ba38bb4da8884990b55de212db77ed6d1c23bf40a06ce1391ed7ba592ba2294f7fb58cc0573834f0cbcec8a840b37d3d8d0363fe66e2a40bec30b06", hex.EncodeToString(vrfProof))

	// Successfull verification
	verified, err := vk.Verify(msg, vrfProof, vrfOut)
	require.Nil(t, err)
	require.True(t, verified)

	// Same key but altered message.
	msgAlt := []byte("[ALT] Proving this is HotPocket")
	vrfOutAlt, vrfProofAlt, err := sk.Prove(msgAlt)
	require.Nil(t, err)

	// Incorrect vrfOut for the original message fails
	verified, err = vk.Verify(msg, vrfProof, vrfOutAlt)
	require.Nil(t, err)
	require.False(t, verified)

	// Incorrect vrfProof for the original message fails
	verified, err = vk.Verify(msg, vrfProofAlt, vrfOut)
	require.Nil(t, err)
	require.False(t, verified)
}
