package utils

import "strings"

// CLEANUP: Only used in one place, so move it into the corresponding module.
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
