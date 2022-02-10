package types

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type PrivateKey ed25519.PrivateKey
type PublicKey ed25519.PublicKey

func GeneratePrivateKey(seed uint32) PrivateKey {
	seedBytes := make([]byte, ed25519.SeedSize)
	binary.BigEndian.PutUint32(seedBytes, seed)
	return PrivateKey(ed25519.NewKeyFromSeed(seedBytes))
}

func AddressFromKey(pubKey PublicKey) string {
	return fmt.Sprintf("%x", sha256.Sum256(pubKey))
}

func (key *PrivateKey) Sign(message []byte) []byte {
	return ed25519.Sign(ed25519.PrivateKey(*key), message)
}

func (key *PrivateKey) Public() PublicKey {
	return PublicKey(ed25519.PrivateKey(*key).Public().(ed25519.PublicKey))
}

func (key *PrivateKey) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", hex.EncodeToString(ed25519.PrivateKey(*key)))), nil
}

func (key *PrivateKey) UnmarshalJSON(b []byte) error {
	s := string(b)
	privKeyFromStrDecoded, err := hex.DecodeString(s[1 : len(s)-1])
	if err != nil {
		return err
	}
	*key = PrivateKey(privKeyFromStrDecoded)
	return nil
}

func (key *PrivateKey) Equal(key2 *PrivateKey) bool {
	return ed25519.PrivateKey(*key).Equal(ed25519.PrivateKey(*key2))
}

func (key *PublicKey) Verify(bytesToVerify, sig []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(*key), bytesToVerify, sig)
}

func (key *PublicKey) UnmarshalJSON(b []byte) error {
	s := string(b)
	pubKeyFromStrDecoded, err := hex.DecodeString(s[1 : len(s)-1])
	if err != nil {
		return err
	}
	*key = PublicKey(pubKeyFromStrDecoded)
	return nil
}

func (key *PublicKey) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", hex.EncodeToString(ed25519.PublicKey(*key)))), nil
}

func (key *PublicKey) Equal(key2 *PublicKey) bool {
	return ed25519.PublicKey(*key).Equal(ed25519.PublicKey(*key2))
}
