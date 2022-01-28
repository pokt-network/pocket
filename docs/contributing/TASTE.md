# Project Wide Code Preferences
This document describes code and patterns that are palatable to the team.

This is the place where default code patterns are encoded. Reasoning for them can be expressed here, especially when discussions arise over time about how justified or not any decision is.

Deviations from these in code contributions should be justified and aren't forbidden: we're not limiting ourselves to a subset of go.

This is a living document. This is supposed to grow as we have more discussions over what constitutes simple, pretty, easy-to-get-right, hard-to-get-wrong, understandable code.

For now, we have these guiding principles:
* Be conservative with dependencies
* Be compatible with the latest stable go version
* Build strings with `fmt.Sprintf`
* Iterate with `range`
* If a meaningful value is used once, document it where used; if used more than once, declare a const and document it there.