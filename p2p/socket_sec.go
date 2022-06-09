package p2p

import (
	"github.com/pokt-network/pocket/p2p/utils"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

const (
	HandshakeSequence = "@pocket::v1::handshake_sequence"
)

func (s *socket) securityFail(err error, reason string) {
	s.logger.Error("Security failure", "err", err, "reason", reason)
	s.reportError(err)
	s.close()
}

// The TLS handshake algorithm to establish encrypted connections
func (s *socket) handshake() {
	// Generate ephemeral Ed25519 keypair to perform a Diffie-Hellman handshake.
	pub, priv := cryptoPocket.GenerateKeyPair()

	// I. Part one:
	// Send our ephemeral public key + a signature of the handshake sequence '.@pocket::v1::handshake_sequence'.

	// I.a Sign the handshake sequence
	//
	// Sign the handshake sequence message with our ephemeral private key.
	// The signature is used to verify that the public key is valid.
	signature, err := priv.Sign([]byte(HandshakeSequence))
	if err != nil {
		s.securityFail(err, "failed to sign handshake sequence")
		return
	}

	// I.b Send the handshake sequence + signature.
	payload := append(pub.Bytes(), signature...)
	if _, err := s.writeChunk(payload, false, 0, 0); err != nil {
		s.securityFail(err, "failed to write handshake sequence")
		return
	}

	// II. Part two:
	// Read from our peer their Ed25519 ephemeral public key and signature of the message '.@pocket::v1::handshake_sequence'.

	// II.a Read the peer's ephemeral public key
	data, err, _ := s.readChunk()
	if err != nil {
		s.securityFail(err, "failed to read handshake sequence")
		return
	}

	// II.a.1: Verify the peer's data size as preliminary check
	if len(data) != cryptoPocket.PublicKeySize+cryptoPocket.SignatureSize {
		s.securityFail(ErrInvalidPublicKeyLen(len(data)), "invalid public key length")
		return
	}

	// Unpack the peer's data
	var peerPublicKey cryptoPocket.PublicKey
	copy(peerPublicKey, data[:cryptoPocket.PublicKeySize])

	// II.b Verify peer's ownership of the public key they sent by verifying the resulting signature.
	isSigAuthentic := peerPublicKey.Verify([]byte(HandshakeSequence), data[cryptoPocket.PublicKeySize:cryptoPocket.PublicKeySize+cryptoPocket.SignatureSize])
	if !isSigAuthentic {
		s.securityFail(ErrInvalidPublicKeySignature, "invalid public key signature")
		return
	}

	// III. Part three:
	// Transform all Ed25519 points to Curve25519 points and perform a Diffie-Hellman handshake to derive a shared key.
	shared, err := utils.DeriveDHSharedKey(priv, peerPublicKey)
	// Use the derived shared key from Diffie-Hellman to encrypt/decrypt all future communications
	// with AES-256 Galois Counter Mode (GCM).

	// Send to our peer our overlay ID.

	// Read and parse from our peer their overlay ID.

	// Validate the peers ownership of the overlay ID.
}
