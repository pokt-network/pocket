package crypto

import (
	"encoding/hex"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// NewLibP2PPrivateKey converts a hex-encoded ed25519d key
// string into a libp2p compatible Private Key.
func NewLibP2PPrivateKey(privateKeyHex string) (crypto.PrivKey, error) {
	privateKeyBz, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, errDecodePrivateKey(err)
	}

	privateKey, err := crypto.PrivKeyUnmarshallers[crypto.Ed25519](privateKeyBz)
	if err != nil {
		return nil, errDecodePrivateKey(err)
	}

	return privateKey, nil
}
