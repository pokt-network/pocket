package vrf

/*
This file is a light wrapper around https://pkg.go.dev/github.com/ProtonMail/go-ecvrf.

It is needed to to achieve more semantic variable naming in the use of HotPocket.
*/

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"io"
	"pocket/shared/crypto"

	"github.com/ProtonMail/go-ecvrf/ecvrf"
)

const (
	VRFOutputSize = sha512.Size // See github.com/ProtonMail/go-ecvrf for details
)

type SecretKey ecvrf.PrivateKey
type VerificationKey ecvrf.PublicKey

// TODO(olshansky): Make these arrays of a specific size rather than slices.
type VRFProof []byte
type VRFOutput []byte

func CreateVRFRandReader(lastBlockHash string, privKey crypto.PrivateKey) (io.Reader, error) {
	if len(lastBlockHash) < crypto.SeedSize/2 {
		return nil, fmt.Errorf("the last block hash must be at least %d bytes in length", crypto.SeedSize/2)
	}

	if privKey == nil {
		return nil, fmt.Errorf("private key cannot be nil")
	}

	privKeySeed := privKey.Seed()[:crypto.SeedSize/2]
	blockHashSeed := lastBlockHash[:crypto.SeedSize/2]

	seed := make([]byte, crypto.SeedSize)
	copy(seed, privKeySeed)
	copy(seed, blockHashSeed)

	return bytes.NewReader(seed), nil
}

// TODO(discuss): Should this return a pointer or a value?
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

// TODO(discuss): Should this return a pointer or a value?
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

// TODO(discuss): Should this return a pointer or a value?
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
