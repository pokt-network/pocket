package common

import "fmt"

type ErrFactory func(msg string, err error) error

func NewErrFactory(preMsg string) ErrFactory {
	return func(msg string, err error) error {
		msgStr := ""
		if msg != "" {
			// NB: with msg - "LibP2P module error: <msg>: <wrapped error>"
			//  without msg - "LibP2P module error: <wrapped error>"
			msgStr = fmt.Sprintf(": %s", msg)
		}
		// TODO: gracefully handle case(s) where msg and err is emtpy.
		return fmt.Errorf("%s%s: %w", preMsg, msgStr, err)
	}
}
