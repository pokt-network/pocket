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
func InlineErrCheck(m dsl.Matcher) {
	m.Match(`$err := $x; if $err != nil { $*_ }`).
		Where(m["err"].Type.Is(`error`)).
		Report(`consider using inline error check: if err := $x; err != nil { $*_ }`).
		Suggest(`if err := $x; err != nil { $*_ }`)
}
