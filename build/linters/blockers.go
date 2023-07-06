//nolint // it's not a Go code file
//go:build !codeanalysis
// +build !codeanalysis

// This file includes our custom linters.
// If you want to add/modify an existing linter, please check out ruleguard's documentation: https://github.com/quasilyte/go-ruleguard#documentation

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

func BlockInThisCommitComment(m dsl.Matcher) {
	if !isFileExcludedForInThisComment(m) {
		m.Match(`//$text`).
			Where(m["text"].Text.Matches(`IN_THIS_COMMIT`)).
			Report(`Don't use IN_THIS_COMMIT in comments`)
	}
}

func BlockInThisPRComment(m dsl.Matcher) {
	if !isFileExcludedForInThisComment(m) {
		m.Match(`//$text`).
			Where(m["text"].Text.Matches(`IN_THIS_PR`)).
			Report(`Don't use IN_THIS_PR in comments`)
	}
}

func isFileExcludedForInThisComment(m dsl.Matcher) bool {
	return m.File().Name == `Makefile` || m.File().Name != `blockers.go`
}
