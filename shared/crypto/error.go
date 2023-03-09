package crypto

import (
	"crypto/ed25519"
	"fmt"
)

const (
	InvalidAddressLenError        = "the address length is not valid"
	InvalidHashLenError           = "the hash length is not valid"
	CreateAddressError            = "an error occurred creating the address"
	InvalidPrivateKeyLenError     = "the private key length is not valid"
	InvalidPrivateKeySeedLenError = "the seed is too short to create a private key"
	CreatePrivateKeyError         = "an error occurred creating the private key"
	InvalidPublicKeyLenError      = "the public key length is not valid"
	CreatePublicKeyError          = "an error occurred creating the public key"
	decodePrivateKeyError         = "decoding private key"
)

func ErrInvalidAddressLen(length int) error {
	return fmt.Errorf("%s, expected length %d, actual length %d", InvalidAddressLenError, AddressLen, length)
}

func ErrInvalidHashLen(length int) error {
	return fmt.Errorf("%s, expected length %d, actual length %d", InvalidHashLenError, SHA3HashLen, length)
}

func ErrCreateAddress(err error) error {
	return fmt.Errorf("%s; %s", CreateAddressError, err.Error())
}

func ErrInvalidPrivateKeyLen(length int) error {
	return fmt.Errorf("%s, expected length %d, actual length %d", InvalidPrivateKeyLenError, ed25519.PrivateKeySize, length)
}

func ErrInvalidPrivateKeySeedLenError(seedLen int) error {
	return fmt.Errorf("%s, expected length %d, actual length %d", InvalidPrivateKeySeedLenError, ed25519.SeedSize, seedLen)
}

func ErrCreatePrivateKey(err error) error {
	return fmt.Errorf("%s; %s", CreatePrivateKeyError, err.Error())
}

func ErrInvalidPublicKeyLen(length int) error {
	return fmt.Errorf("%s, expected length %d, actual length: %d", InvalidPublicKeyLenError, ed25519.PublicKeySize, length)
}

func ErrCreatePublicKey(err error) error {
	return fmt.Errorf("%s; %s", CreatePublicKeyError, err.Error())
}

func errDecodePrivateKey(err error) error {
	return fmt.Errorf("%s; %w", decodePrivateKeyError, err)
}
