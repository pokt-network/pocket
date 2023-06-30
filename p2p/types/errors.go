package types

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownPeer = errors.New("unknown peer")
)

func ErrUnknownEventType(msg any) error {
	return fmt.Errorf("unknown event type: %v", msg)
}
