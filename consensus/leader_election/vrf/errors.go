package vrf

import (
	"errors"
	"fmt"
)

const (
	NilPrivateKeyError      = "private key cannot be nil"
	BadStateHashLengthError = "the previous block hash must be at least %d bytes in length"
)

var (
	ErrNilPrivateKey = errors.New(NilPrivateKeyError)
)

func ErrBadStateHashLength(seedSize int) error {
	return fmt.Errorf(BadStateHashLengthError, seedSize)
}
