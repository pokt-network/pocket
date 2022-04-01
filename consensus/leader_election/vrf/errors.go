package vrf

import (
	"errors"
	"fmt"
)

const (
	NilPrivateKeyError    = "private key cannot be nil"
	BadAppHashLengthError = "the last block hash must be at least %d bytes in length"
)

var (
	ErrNilPrivateKey = errors.New(NilPrivateKeyError)
)

func ErrBadAppHashLength(seedSize int) error {
	return fmt.Errorf(BadAppHashLengthError, seedSize)
}
