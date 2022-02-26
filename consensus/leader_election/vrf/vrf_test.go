package vrf

import (
	"encoding/hex"
	"pocket/shared/crypto"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVRFKeygenWithoutSeed(t *testing.T) {
	sk, vk, err := GenerateVRFKeys(nil)
	require.Nil(t, err)
	require.NotNil(t, sk)
	require.NotNil(t, vk)
}

func TestVRFKeygenWithSeed(t *testing.T) {
	seed := "Olshansky wonders if anyone is ever going to read this code"
	require.GreaterOrEqual(t, len(seed), crypto.SeedSize/2)

	privKey, err := crypto.NewPrivateKeyFromSeed([]byte(seed))
	require.Nil(t, err)
	lastBlockHash := seed

	reader, err := CreateVRFRandReader(lastBlockHash, privKey)
	require.Nil(t, err)
	require.NotNil(t, reader)

	sk, vk, err := GenerateVRFKeys(reader)
	require.Nil(t, err)

	require.Equal(t, "4f6c7368616e736b7920776f6e64657200000000000000000000000000000000c8491df826eccf7557467c74f7a93bb324a15efd2359dc27e3eba940127ff8a2", hex.EncodeToString(sk.Bytes()))
	require.Equal(t, "c8491df826eccf7557467c74f7a93bb324a15efd2359dc27e3eba940127ff8a2", hex.EncodeToString(vk.Bytes()))
}

func TestVRFKeygenProveAndVerify(t *testing.T) {
	msg := []byte("HotPocket: Gotta prove it like it's hot.")

	sk, vk, err := GenerateVRFKeys(nil)
	require.Nil(t, err)

	vrfOut, vrfProof, err := sk.Prove(msg)
	require.Nil(t, err)

	// Successfull verification
	verified, err := vk.Verify(msg, vrfProof, vrfOut)
	require.Nil(t, err)
	require.True(t, verified)

	// Same key but altered message.
	msgAlt := []byte("To paraphrase the infamous Katy Perry: If you're hot, but then you're cold, you can't prove it like it's hot.")
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

func TestVRFKeygenProveAndVerifyWithSeed(t *testing.T) {
	seed := "So you read the code for the first test, but did you get to the second?"
	require.GreaterOrEqual(t, len(seed), crypto.SeedSize/2)

	privKey, err := crypto.NewPrivateKeyFromSeed([]byte(seed))
	require.Nil(t, err)
	lastBlockHash := seed

	reader, err := CreateVRFRandReader(lastBlockHash, privKey)
	require.Nil(t, err)
	require.NotNil(t, reader)

	msg := []byte("Proving this is HotPocket")

	sk, vk, err := GenerateVRFKeys(reader)
	require.Nil(t, err)

	vrfOut, vrfProof, err := sk.Prove(msg)
	require.Nil(t, err)
	require.Equal(t, "d4c95d83e26323ec6e86801d810071aefbface10ac59c250e58096f18a72b56c8d9166cfc8252bbb80def11f438d5ce484373f718261555b59eb6f6d9af9370a", hex.EncodeToString(vrfOut))
	require.Equal(t, "3d277cbd2d7ecde326e2cd3cf3d7787997c52fe7bf98c18e8417f4b5e2e7d78368ef28822f2e4b3d806ed4e5cbc492c67d9bcb86b09c9c49978712041d2ffd7aa433dc7a326362fe70657a66af3a220d", hex.EncodeToString(vrfProof))

	// Successfull verification
	verified, err := vk.Verify(msg, vrfProof, vrfOut)
	require.Nil(t, err)
	require.True(t, verified)

	// Same key but altered message.
	msgAlt := []byte("No one wants to eat a cold HotPocket")
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
