package vrf

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha512"
	"fmt"
	"io"

	"github.com/ProtonMail/go-ecvrf/ecvrf"
)

/*
This file is a light wrapper around https://pkg.go.dev/github.com/ProtonMail/go-ecvrf/
to achieve more semantic variable naming throughout HotPocket.
*/

const VRFOutputSize = sha512.Size // github.com/ProtonMail/go-ecvrf

type SecretKey ecvrf.PrivateKey
type VerificationKey ecvrf.PublicKey

// TODO: Make these arrays of a specific size rather than slices.
type VRFProof []byte
type VRFOutput []byte

func CreateVRFRandReader(lastBlockHash string, privKey *ed25519.PrivateKey) (io.Reader, error) {
	if len(lastBlockHash) < ed25519.SeedSize/2 {
		return nil, fmt.Errorf("the last block hash must be at least %d bytes in length", ed25519.SeedSize/2)
	}

	if privKey == nil {
		return nil, fmt.Errorf("private key cannot be nil")
	}

	seed := make([]byte, ed25519.SeedSize)
	privKeySeed := privKey.Seed()
	copy(seed, privKeySeed[:ed25519.SeedSize/2])
	copy(seed, privKeySeed[ed25519.SeedSize/2:])

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

func (key *VerificationKey) Bytes() []byte {
	return (*ecvrf.PublicKey)(key).Bytes()
}

func (key *VerificationKey) Verify(msg, vrfProof VRFProof, vrfOut VRFOutput) (verified bool, err error) {
	verified, vrf, err := (*ecvrf.PublicKey)(key).Verify(msg, vrfProof)
	verified = verified && bytes.Equal(vrf, vrfOut)
	return
}

func (key *SecretKey) Bytes() []byte {
	return (*ecvrf.PrivateKey)(key).Bytes()
}

func (key *SecretKey) Prove(msg []byte) (vrf VRFOutput, proof VRFProof, err error) {
	vrf, proof, err = (*ecvrf.PrivateKey)(key).Prove(msg)
	return
}

func (key *SecretKey) VerificationKey() (vrf *VerificationKey, err error) {
	verificationKey, err := (*ecvrf.PrivateKey)(key).Public()
	if err != nil {
		return nil, err
	}

	return (*VerificationKey)(verificationKey), nil
}
