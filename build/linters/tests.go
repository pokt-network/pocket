//nolint // it's not a Go code file
//go:build !codeanalysis
// +build !codeanalysis

// This file includes our custom linters.
// If you want to add/modify an existing linter, please check out ruleguard's documentation: https://github.com/quasilyte/go-ruleguard#documentation

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

// This is a custom linter that checks ensures a use of require.Equal
func EqualInsteadOfTrue(m dsl.Matcher) {
	m.Match(`require.True($t, $x == $y, $*args)`).
		Suggest(`require.Equal($t, $x, $y, $args)`).
		Report(`use require.Equal instead of require.True`)
}
