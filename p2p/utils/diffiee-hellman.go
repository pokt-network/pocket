package utils

import (
	"github.com/pokt-network/pocket/shared/crypto"
	"golang.org/x/crypto/curve25519"
)

func DeriveDHSharedKey(privateKey crypto.Ed25519PrivateKey, publicKey crypto.Ed25519PrivateKey) ([]byte, error) {
	curvePrivKey := privateKey.ToCurve25519()
	curvePubKey := publicKey.ToCurve25519()

	shared, err := curve25519.X25519(curvePrivKey, curvePubKey)
	if err != nil {
		return nil, ErrSharedKeyCreationFailed(err)
	}

	return shared, nil
}
