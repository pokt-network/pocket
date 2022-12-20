// NOTE: At the time of implementation, the VRF library is only used with consensus, but may
// be extracted into shared/crypto when/if it will be needed elsewhere.
package vrf

// This file is a light wrapper around https://pkg.go.dev/github.com/ProtonMail/go-ecvrf.
// It is needed to to achieve more semantic variable naming in the use of HotPocket.

import (
	"bytes"
	"crypto/sha512"
	"io"

	"github.com/pokt-network/pocket/internal/shared/crypto"

	"github.com/ProtonMail/go-ecvrf/ecvrf"
)

const (
	VRFOutputSize = sha512.Size // See github.com/ProtonMail/go-ecvrf for details
)

type (
	SecretKey       ecvrf.PrivateKey
	VerificationKey ecvrf.PublicKey
)

// These are slices rather than arrays in order to more easily comply with the underlying `go-ecvrf/ecvrf` library.
type (
	VRFProof  []byte // A proof to verify that VRFOutput belongs to a certain publicKey.
	VRFOutput []byte // Uniformally distributed output that can be normalized to be used in a binomial distribution.
)

func CreateVRFRandReader(lastBlockHash string, privKey crypto.PrivateKey) (io.Reader, error) {
	if privKey == nil {
		return nil, ErrNilPrivateKey
	}

	if len(lastBlockHash) < crypto.SeedSize/2 {
		return nil, ErrBadAppHashLength(crypto.SeedSize)
	}

	privKeySeed := privKey.Seed()[:crypto.SeedSize/2]
	blockHashSeed := lastBlockHash[:crypto.SeedSize/2]

	seed := make([]byte, crypto.SeedSize)
	copy(seed, privKeySeed)
	copy(seed, blockHashSeed)

	return bytes.NewReader(seed), nil
}

func GenerateVRFKeys(reader io.Reader) (*SecretKey, *VerificationKey, error) {
	privateKey, err := ecvrf.GenerateKey(reader)
	if err != nil {
		return nil, nil, err
	}

	publicKey, err := privateKey.Public()
	if err != nil {
		return nil, nil, err
	}

	return (*SecretKey)(privateKey), (*VerificationKey)(publicKey), nil
}

func VerificationKeyFromBytes(data []byte) (*VerificationKey, error) {
	key, err := ecvrf.NewPublicKey(data)
	if err != nil {
		return nil, err
	}
	return (*VerificationKey)(key), nil
}

func (key *VerificationKey) Verify(msg []byte, vrfProof VRFProof, vrfOut VRFOutput) (verified bool, err error) {
	verified, vrf, err := (*ecvrf.PublicKey)(key).Verify(msg, vrfProof)
	verified = verified && bytes.Equal(vrf, vrfOut)
	return
}

func (key *VerificationKey) Bytes() []byte {
	return (*ecvrf.PublicKey)(key).Bytes()
}

func (key *SecretKey) Prove(msg []byte) (vrf VRFOutput, proof VRFProof, err error) {
	vrf, proof, err = (*ecvrf.PrivateKey)(key).Prove(msg)
	return
}

func (key *SecretKey) VerificationKey() (*VerificationKey, error) {
	verificationKey, err := (*ecvrf.PrivateKey)(key).Public()
	if err != nil {
		return nil, err
	}

	return (*VerificationKey)(verificationKey), nil
}

func (key *SecretKey) Bytes() []byte {
	return (*ecvrf.PrivateKey)(key).Bytes()
}
