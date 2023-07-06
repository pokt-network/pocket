//nolint // it's not a Go code file
//go:build !codeanalysis
// +build !codeanalysis

// This file includes our custom linters.
// If you want to add/modify an existing linter, please check out ruleguard's documentation: https://github.com/quasilyte/go-ruleguard#documentation

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

// Blocks merge if IN_THIS_COMMIT comments are present
func BlockInThisCommitComment(m dsl.Matcher) {
	m.Match(`//$text`).
		Where(isFileExcludedForInThisComment(m)).
		Where(m["text"].Text.Matches(`IN_THIS_COMMIT`)).
		Report(`IN_THIS_COMMIT comment must be addressed before merging to main`)

}

// Blocks merge if IN_THIS_PR comments are present
func BlockInThisPRComment(m dsl.Matcher) {
	m.Match(`//$text`).
		Where(isFileExcludedForInThisComment(m)).
		Where(m["text"].Text.Matches(`IN_THIS_PR`)).
		Report(`IN_THIS_PR comment must be addressed before merging to main`)
}

func isFileExcludedForInThisComment(m dsl.Matcher) bool {
	return m.File().Name == `Makefile` || m.File().Name != `blockers.go`
}
