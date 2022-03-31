package crypto

import (
	"crypto/ed25519"
	"errors"
	"fmt"
)

const (
	InvalidAddressLenError    = "the address length is not valid"
	InvalidHashLenError       = "the hash length is not valid"
	CreateAddressError        = "an error occurred creating the address"
	InvalidPrivateKeyLenError = "the private key length is not valid"
	CreatePrivateKeyError     = "an error occurred creating the private key"
	InvalidPublicKeyLenError  = "the public key length is not valid"
	CreatePublicKeyError      = "an error occurred creating the private key"
)

func ErrInvalidAddressLen(len int) error {
	return fmt.Errorf("%s, expected length %d, actual length %d", InvalidAddressLenError, AddressLen, len)
}

func ErrInvalidHashLen(len int) error {
	return fmt.Errorf("%s, expected length %d, actual length %d", InvalidHashLenError, SHA3HashLen, len)
}

func ErrCreateAddress(err error) error {
	return fmt.Errorf("%s; %s", CreateAddressError, err.Error())
}

func ErrInvalidPrivateKeyLen(len int) error {
	return fmt.Errorf("%s, expected length %d, actual length %d", InvalidPrivateKeyLenError, ed25519.PrivateKeySize, len)
}

func ErrCreatePrivateKey(err error) error {
	return fmt.Errorf("%s; %s", CreatePrivateKeyError, err.Error())
}

func ErrInvalidPublicKeyLen(len int) error {
	return fmt.Errorf("%s, expected length %d, actual length: %d", InvalidPublicKeyLenError, ed25519.PrivateKeySize, len)
}

func ErrCreatePublicKey(err error) error {
	return fmt.Errorf("%s; %s", CreatePublicKeyError, err.Error())
	return errors.New(fmt.Sprintf("%s, expected length %d", InvalidAddressLenError, AddressLen))
}
