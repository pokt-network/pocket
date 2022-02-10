package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeygenCodec(t *testing.T) {
	seed := uint32(1)
	privKey := GeneratePrivateKey(seed)
	pubKey := privKey.Public()

	privKeyStr := hex.EncodeToString(privKey)
	pubKeyStr := hex.EncodeToString(pubKey)
	address := AddressFromKey(pubKey)

	require.Equal(t, "0000000100000000000000000000000000000000000000000000000000000000b1f804dabc68274c1233995c5a9119b56935bcdd83b7de07ec726dcedc4e9ce7", privKeyStr)
	require.Equal(t, "b1f804dabc68274c1233995c5a9119b56935bcdd83b7de07ec726dcedc4e9ce7", pubKeyStr)
	require.Equal(t, "71f8be163036c0da94f188bb817d77691869ccff5932059f3c398f2fb92fa08b", address)

	privKeyFromStrDecoded, _ := hex.DecodeString(privKeyStr)
	privKeyFromStr := PrivateKey(privKeyFromStrDecoded)

	pubKeyFromStrDecoded, _ := hex.DecodeString(pubKeyStr)
	pubKeyFromStr := PublicKey(pubKeyFromStrDecoded)

	require.True(t, privKeyFromStr.Equal(&privKey))
	require.True(t, pubKeyFromStr.Equal(&pubKey))
}

func TestKeygenMarshal(t *testing.T) {
	seed := uint32(1)

	// Private key
	privKey := GeneratePrivateKey(seed)
	privKeyMarshalled, err := json.Marshal(&privKey)
	require.NoError(t, err)

	privKeyUnmarshalled := PrivateKey{}
	err = json.Unmarshal(privKeyMarshalled, &privKeyUnmarshalled)
	require.NoError(t, err)

	require.Equal(t, privKeyUnmarshalled, privKey)

	// Public key
	pubKey := privKey.Public()
	pubKeyMarshalled, err := json.Marshal(&pubKey)
	require.NoError(t, err)

	pubKeyUnmarshalled := PublicKey{}
	err = json.Unmarshal(pubKeyMarshalled, &pubKeyUnmarshalled)
	require.NoError(t, err)

	require.Equal(t, privKeyUnmarshalled, privKey)
}

func TestGenerateKeysUtility(t *testing.T) {
	for i := uint32(1); i < 5; i++ {
		privKey := GeneratePrivateKey(i)
		pubKey := privKey.Public()

		privKeyStr := hex.EncodeToString(privKey)
		pubKeyStr := hex.EncodeToString(pubKey)
		address := AddressFromKey(pubKey)

		fmt.Printf("NodeId %d:\n\tAddress: %s;\n\tPubKey: %s;\n\tPrivKey: %s\n", i, address, pubKeyStr, privKeyStr)
	}
}
