//go:build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/sh"
)

const (
	versionStringEnvVarName = "POCKET_VERSION_STRING"
	pocketPackage           = "github.com/pokt-network/pocket/cmd/pocket"
	ldflags                 = "-X main.version=$" + versionStringEnvVarName
)

var env = map[string]string{}

func setup() {
	env["GO111MODULE"] = "on"

	fmt.Println("Getting branch and commit information")
	branch, branchError := sh.Output("git", "branch", "--show-current")
	hash, hashError := sh.Output("git", "rev-parse", "--short", "HEAD")
	_, dirtyError := sh.Output("git", "diff", "--quiet")

	versionNumber := "UNKNOWN"

	// git repo invariant: we have a branch and a hash whenever version is known
	if branchError == nil && hashError == nil {
		versionNumber = fmt.Sprintf("0.0.0-%s/%s", branch, hash)
		if dirtyError != nil {
			versionNumber += "+dirty"
		}
	}
	env[versionStringEnvVarName] = versionNumber
}

// Builds the pocket executable and puts it in ./build
func Build() error {
	setup()
	return sh.RunWith(env, "go", "build", "-o", "bin/", "-ldflags", ldflags, pocketPackage)
}

// Builds the pocket executable with race detection enabled. Not for production.
func BuildRace() error {
	setup()
	env[versionStringEnvVarName] += "+race"
	return sh.RunWith(env, "go", "build", "-o", "bin/", "-ldflags", ldflags, "-race", pocketPackage)
}

// Installs the pocket executable in the target used by go install.
func Install() error {
	setup()
	return sh.RunWith(env, "go", "install", "-ldflags", ldflags, pocketPackage)
}

// Uninstalls the pocket executable if previously added with the install target.
func Uninstall() error {
	return sh.Run("go", "clean", "-i", pocketPackage)
}
