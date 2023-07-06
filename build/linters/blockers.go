//nolint // it's not a Go code file
//go:build !codeanalysis
// +build !codeanalysis

// This file includes our custom linters.
// If you want to add/modify an existing linter, please check out ruleguard's documentation: https://github.com/quasilyte/go-ruleguard#documentation

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

// Blocks merge if _IN_THIS_ comments are present
func BlockInThisCommitPRComment(m dsl.Matcher) {
	m.MatchComment(`//.*_IN_THIS_.*`).
		Where(!m.File().Name.Matches(`Makefile`) && !m.File().Name.Matches(`blockers.go`)).
		Report("'_IN_THIS_' comments must be addressed before merging to main")
}
