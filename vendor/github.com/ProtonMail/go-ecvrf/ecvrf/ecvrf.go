// Package ecvrf implements ECVRF-EDWARDS25519-SHA512-TAI, a verifiable random
// function described in draft-irtf-cfrg-vrf-10.
// This VRF uses the Edwards form of Curve25519, SHA512 and the try-and-increment
// hash-to-curve function.
// See: https://datatracker.ietf.org/doc/draft-irtf-cfrg-vrf/
package ecvrf

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"errors"
	"io"

	"filippo.io/edwards25519"
)

const (
	scalarSize       = ed25519.SeedSize
	pointSize        = ed25519.PublicKeySize
	intermediateSize = 16
	PublicKeySize    = pointSize
	PrivateKeySize   = scalarSize + pointSize
	ProofSize        = pointSize + intermediateSize + scalarSize
	suiteID          = 0x03
)

// PrivateKey contains the VRF private key, a standard ed25519 key and the precomputed secret scalar.
type PrivateKey struct {
	sk, pk []byte
	x      *edwards25519.Scalar
}

// PublicKey contains the VRF public key, the canonical representation of a point on the ed25519 curve.
type PublicKey struct {
	pk []byte
}

// GenerateKey creates a public/private key pair using rnd for randomness.
// If rnd is nil, crypto/rand is used.
func GenerateKey(rnd io.Reader) (sk *PrivateKey, err error) {
	if rnd == nil {
		rnd = rand.Reader
	}

	seed := make([]byte, scalarSize)
	if _, err := io.ReadFull(rnd, seed); err != nil {
		return nil, err
	}

	// Generate the private key, the secret scalar, and the public key
	// according to Section 5.1.5 of RFC8032 and cache the values.
	h := sha512.Sum512(seed)
	s, err := edwards25519.NewScalar().SetBytesWithClamping(h[:scalarSize])
	if err != nil {
		return nil, err
	}

	A := (&edwards25519.Point{}).ScalarBaseMult(s)
	return &PrivateKey{seed, A.Bytes(), s}, err
}

// NewPrivateKey generates a PrivateKey object from a standard RFC8032
// Ed25519 64-byte private key.
func NewPrivateKey(skBytes []byte) (sk *PrivateKey, err error) {
	if len(skBytes) != PrivateKeySize {
		return nil, errors.New("ecvrf: bad private key size")
	}

	// Generate the secret scalar according to Section 5.1.5 of RFC8032.
	h := sha512.Sum512(skBytes[:scalarSize])
	s, err := edwards25519.NewScalar().SetBytesWithClamping(h[:scalarSize])
	if err != nil {
		return nil, err
	}

	return &PrivateKey{skBytes[:scalarSize], skBytes[scalarSize:], s}, err
}

// Public extracts the public VRF key from the underlying private-key.
func (sk *PrivateKey) Public() (*PublicKey, error) {
	return NewPublicKey(sk.pk)
}

// Bytes serialises the private VRF key in a bytearray.
func (sk *PrivateKey) Bytes() []byte {
	buf := make([]byte, PrivateKeySize)
	copy(buf, sk.sk)
	copy(buf[scalarSize:], sk.pk)
	return buf
}

// NewPublicKey generates a PublicKey object from a standard RFC8032
// Ed25519 32-byte public key.
func NewPublicKey(pkBytes []byte) (*PublicKey, error) {
	return &PublicKey{pkBytes}, nil
}

// Bytes serialises the private VRF key in a bytearray.
func (pk *PublicKey) Bytes() []byte {
	return pk.pk
}

// Prove returns a proof such that Verify(pk, message, vrf, proof) == true
// for a given message and public key pair sk/pk.
// This function is defined in section 5.1 of draft-irtf-cfrg-vrf-10.
func (sk *PrivateKey) Prove(message []byte) (vrf, proof []byte, err error) {
	// Step 1 is done in key generation/parsing
	h, err := hashToCurveTAI(sk.pk, message)
	if err != nil {
		return nil, nil, err
	}

	gamma := (&edwards25519.Point{}).ScalarMult(sk.x, h)
	kHash := generateNonceHash(sk.sk, h.Bytes())
	k, err := edwards25519.NewScalar().SetUniformBytes(kHash)
	if err != nil {
		return nil, nil, err
	}

	c := hashPoints(
		h,
		gamma,
		(&edwards25519.Point{}).ScalarBaseMult(k),
		(&edwards25519.Point{}).ScalarMult(k, h),
	)

	// append 16 zeroes to c to convert it to a scalar
	cScal, err := cToScalar(c)
	if err != nil {
		return nil, nil, err
	}

	// s = (k + c*x) mod q
	s := edwards25519.NewScalar().Add(k, edwards25519.NewScalar().Multiply(cScal, sk.x))

	proof = make([]byte, ProofSize)
	copy(proof, gamma.Bytes())
	copy(proof[pointSize:], c)
	copy(proof[pointSize+intermediateSize:], s.Bytes())

	return proofToHash(gamma), proof, nil
}

// Verify verifies that the given proof matches the message and the public
// key pk. When true it also returns the expected VRF string.
// This function is defined in section 5.3 of draft-irtf-cfrg-vrf-10.
func (pk *PublicKey) Verify(message, proof []byte) (verified bool, vrf []byte, err error) {
	if len(proof) != ProofSize {
		return false, nil, errors.New("ecvrf: bad proof length")
	}

	y, err := (&edwards25519.Point{}).SetBytes(pk.pk)
	if err != nil {
		return false, nil, err
	}

	gamma, err := (&edwards25519.Point{}).SetBytes(proof[:pointSize])
	if err != nil {
		return false, nil, err
	}

	c, err := cToScalar(proof[pointSize : pointSize+intermediateSize])
	if err != nil {
		return false, nil, err
	}

	s, err := edwards25519.NewScalar().SetCanonicalBytes(proof[pointSize+intermediateSize:])
	if err != nil {
		return false, nil, err
	}

	h, err := hashToCurveTAI(pk.pk, message)
	if err != nil {
		return false, nil, err
	}

	// U = s*B - c*Y
	u := (&edwards25519.Point{}).Subtract(
		(&edwards25519.Point{}).ScalarBaseMult(s),
		(&edwards25519.Point{}).ScalarMult(c, y),
	)

	// V = s*H - c*Gamma
	v := (&edwards25519.Point{}).Subtract(
		(&edwards25519.Point{}).ScalarMult(s, h),
		(&edwards25519.Point{}).ScalarMult(c, gamma),
	)

	// If c and c' are different
	if subtle.ConstantTimeCompare(hashPoints(h, gamma, u, v), proof[pointSize:pointSize+intermediateSize]) == 0 {
		return false, nil, nil
	}

	return true, proofToHash(gamma), nil
}

// -- internal functions --

// Step 5.1.2 of draft-irtf-cfrg-vrf-10 implemented as defined by the section
// 5.4.1.1, ECVRF_hash_to_curve_try_and_increment.
func hashToCurveTAI(pk, alpha []byte) (*edwards25519.Point, error) {
	// CTR needs to be encoded in a string of length 1, therefore at most 255
	for ctr := 0; ctr < 256; ctr++ {
		h := sha512.New()
		h.Write([]byte{suiteID, 0x01})
		h.Write(pk)
		h.Write(alpha)
		h.Write([]byte{uint8(ctr), 0x00})

		p, err := (&edwards25519.Point{}).SetBytes(h.Sum(nil)[:scalarSize])
		if err == nil && p.Equal(edwards25519.NewIdentityPoint()) == 0 {
			return (&edwards25519.Point{}).MultByCofactor(p), nil
		}
	}

	// Abort - too many CTR attempts
	return nil, errors.New("ecvrf: unable to find suitable ctr value")
}

// generateNonceHash implements step 5.1.5 of draft-irtf-cfrg-vrf-10 as defined by the section
// 5.4.2.2, ECVRF_nonce_generation_RFC8032.
func generateNonceHash(sk, h []byte) []byte {
	skHash := sha512.New()
	skHash.Write(sk)

	nonceHash := sha512.New()
	nonceHash.Write(skHash.Sum(nil)[scalarSize:])
	nonceHash.Write(h)

	return nonceHash.Sum(nil)
}

// hashPoints implements step 5.1.6 of draft-irtf-cfrg-vrf-10 as defined by the section
// 5.4.3, ECVRF_hash_points.
func hashPoints(points ...*edwards25519.Point) []byte {
	h := sha512.New()
	h.Write([]byte{suiteID, 0x02})
	for _, point := range points {
		h.Write(point.Bytes())
	}
	h.Write([]byte{0x00})

	return h.Sum(nil)[:intermediateSize]
}

// proofToHash implements section 5.2 of draft-irtf-cfrg-vrf-10.
func proofToHash(gamma *edwards25519.Point) []byte {
	h := sha512.New()
	gammaC := (&edwards25519.Point{}).MultByCofactor(gamma)
	h.Write([]byte{suiteID, 0x03})
	h.Write(gammaC.Bytes())
	h.Write([]byte{0x00})

	return h.Sum(nil)
}

// cToScalar transforms the 16-byte c into an ed25519 scalar.
func cToScalar(c []byte) (*edwards25519.Scalar, error) {
	cRaw := make([]byte, 32)
	copy(cRaw, c)
	return edwards25519.NewScalar().SetCanonicalBytes(cRaw)
}
