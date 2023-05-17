//go:build debug

package main

import (
	_ "github.com/pokt-network/pocket/app/client/keybase/debug"
)

// This file serves as a feature flag based on the build tag "debug".
// When the build tag "debug" is present, the init() function in keystore.go is triggered via the anonymous import above.
// This functionality is intended for debugging purposes only.

// Additional debug functionality in the client CLI could be included here.
// For example, logging or error handling utilities that are useful for developers when debugging their applications.
// To run the client CLI with the "debug" build tag, run the following command: go run -tags=debug app/client/*.go
