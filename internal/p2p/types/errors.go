package types

import "fmt"

func ErrUnknownEventType(msg interface{}) error {
	return fmt.Errorf("unknown event type: %v", msg)
}
