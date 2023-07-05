package types

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidNonce = errors.New("invalid nonce")
)

func ErrUnknownEventType(msg any) error {
	return fmt.Errorf("unknown event type: %v", msg)
}
