package common

import "fmt"

type ErrFactory func(msg string, err error) error

func NewErrFactory(preMsg string) ErrFactory {
	return func(msg string, err error) error {
		msgStr := ""
		if msg != "" {
			// NB: with msg - "<preMsg>: <msg>: <wrapped error>"
			//  without msg - "<preMsg>: <wrapped error>"
			msgStr = fmt.Sprintf(": %s", msg)
		}
		// TODO: gracefully handle case(s) where preMsg, msg, and/or err are empty.
		return fmt.Errorf("%s%s: %w", preMsg, msgStr, err)
	}
}
