package types

import "fmt"

func ErrUnknownEventType(msg any) error {
	return fmt.Errorf("unknown event type: %v", msg)
}
