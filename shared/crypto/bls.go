package bls

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/stretchr/testify/require"
)

var (
	message              = GenRandomBytes(5000)
	secretKey, publicKey = bls.GenerateRandomKey()
	signature            = secretKey.Sign(message)
	pubBytes             = publicKey.Marshal()
	sigBytes             = signature.Marshal()
)

func Test_VerifySignature(t *testing.T) {
	require.True(t, signature.Verify(publicKey, message))
}

