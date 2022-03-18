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
)

func ErrInvalidAddressLen() error {
	return fmt.Errorf("%s, expected length %d", InvalidAddressLenError, AddressLen)
}

func ErrInvalidHashLen() error {
	return fmt.Errorf("%s, expected length %d", InvalidHashLenError, SHA3HashLen)
}

func ErrCreateAddress(err error) error {
	return fmt.Errorf("%s; %s", CreateAddressError, err.Error())
}

func ErrInvalidPrivateKeyLen() error {
	return fmt.Errorf("%s, expected length %d", InvalidPrivateKeyLenError, ed25519.PrivateKeySize)
}

func ErrInvalidPrivateKeySeedLenError(seedLen int) error {
	return fmt.Errorf("%s, expected length %d, actual length %d", InvalidPrivateKeySeedLenError, ed25519.SeedSize, seedLen)
}

func ErrCreatePrivateKey(err error) error {
	return fmt.Errorf("%s; %s", CreatePrivateKeyError, err.Error())
}

func ErrInvalidPublicKeyLen() error {
	return fmt.Errorf("%s, expected length %d", InvalidPublicKeyLenError, ed25519.PrivateKeySize)
}

func ErrCreatePublicKey(err error) error {
	return fmt.Errorf("%s; %s", CreatePublicKeyError, err.Error())
}
