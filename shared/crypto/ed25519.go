package crypto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const (
	AddressLen = 20
)

type (
	Ed25519PublicKey  ed25519.PublicKey
	Ed25519PrivateKey ed25519.PrivateKey
)

var (
	PublicKeyLen  = ed25519.PublicKeySize
	PrivateKeyLen = ed25519.PrivateKeySize
)

func NewAddress(hexString string) (Address, error) {
	bz, err := hex.DecodeString(hexString)
	if err != nil {
		return bz, ErrCreateAddress(err)
	}
	return NewAddressFromBytes(bz)
}

func NewAddressFromBytes(bz []byte) (Address, error) {
	if len(bz) != AddressLen {
		return bz, ErrInvalidAddressLen()
	}
	return bz, nil
}

func (a Address) String() string {
	return hex.EncodeToString(a)
}

func NewPrivateKey(hexString string) (PrivateKey, error) {
	bz, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, ErrCreatePrivateKey(err)
	}
	return NewPrivateKeyFromBytes(bz)
}

func GeneratePrivateKey() (PrivateKey, error) {
	_, pk, err := ed25519.GenerateKey(nil)
	return Ed25519PrivateKey(pk), err
}

func NewPrivateKeyFromBytes(b []byte) (PrivateKey, error) {
	if len(b) != ed25519.PrivateKeySize {
		return nil, ErrInvalidPrivateKeyLen()
	}
	return Ed25519PrivateKey(b), nil
}

var _ PrivateKey = Ed25519PrivateKey{}

func (priv Ed25519PrivateKey) Bytes() []byte {
	return priv
}

func (priv Ed25519PrivateKey) String() string {
	return hex.EncodeToString(priv.Bytes())
}

func (priv Ed25519PrivateKey) Equals(other PrivateKey) bool {
	return ed25519.PrivateKey(priv).Equal(ed25519.PrivateKey(other.(Ed25519PrivateKey)))
}

func (priv Ed25519PrivateKey) PublicKey() PublicKey {
	pubKey := ed25519.PrivateKey(priv).Public()
	return Ed25519PublicKey(pubKey.(ed25519.PublicKey))
}

func (priv Ed25519PrivateKey) Address() Address {
	publicKey := priv.PublicKey()
	return publicKey.Address()
}

func (priv Ed25519PrivateKey) Sign(msg []byte) ([]byte, error) {
	return ed25519.Sign(ed25519.PrivateKey(priv), msg), nil
}

func (priv Ed25519PrivateKey) Size() int {
	return ed25519.PrivateKeySize
}

func (priv *Ed25519PrivateKey) UnmarshalJSON(data []byte) error {
	var privateKey string
	err := json.Unmarshal(data, &privateKey)
	if err != nil {
		return err
	}
	*priv = []byte(privateKey)
	return nil
}

var _ PublicKey = Ed25519PublicKey{}

func NewPublicKey(hexString string) (PublicKey, error) {
	bz, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, ErrCreatePublicKey(err)
	}
	return NewPublicKeyFromBytes(bz)
}

func NewPublicKeyFromBytes(b []byte) (PublicKey, error) {
	if len(b) != ed25519.PublicKeySize {
		return nil, ErrInvalidPublicKeyLen()
	}
	return Ed25519PublicKey(b), nil
}

func (pub Ed25519PublicKey) Bytes() []byte {
	return pub
}

func (pub Ed25519PublicKey) String() string {
	return hex.EncodeToString(pub.Bytes())
}

func (pub Ed25519PublicKey) Address() Address {
	hash := sha256.Sum256(pub[:])
	return hash[:AddressLen]
}

func (pub Ed25519PublicKey) Equals(other PublicKey) bool {
	return ed25519.PublicKey(pub).Equal(ed25519.PublicKey(other.(Ed25519PublicKey)))
}

func (pub Ed25519PublicKey) VerifyBytes(msg []byte, sig []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(pub), msg, sig)
}

func (pub Ed25519PublicKey) Size() int {
	return ed25519.PublicKeySize
}

func GeneratePublicKey() (PublicKey, error) {
	pk, err := GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	return pk.PublicKey(), nil
}

func GenerateAddress() (Address, error) {
	pk, err := GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	return pk.Address(), nil
}

func (pub *Ed25519PublicKey) UnmarshalJSON(data []byte) error {
	var publicKey string
	err := json.Unmarshal(data, &publicKey)
	if err != nil {
		return err
	}
	*pub = []byte(publicKey)
	return nil
}
