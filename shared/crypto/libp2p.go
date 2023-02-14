package crypto

import (
	"encoding/hex"
	"github.com/libp2p/go-libp2p/core/crypto"
)

// NewLibP2PPrivateKey converts a hex-encoded ed25519d key
// string into a libp2p compatible Private Key.
func NewLibP2PPrivateKey(hexString string) (crypto.PrivKey, error) {
	keyBytes, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, ErrCreatePrivateKey(err)
	}

	privateKey, err := crypto.PrivKeyUnmarshallers[crypto.Ed25519](keyBytes)
	if err != nil {
		return nil, ErrCreatePublicKey(err)
	}

	return privateKey, nil
}
