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

func init() {
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

func Build() error {
	return sh.RunWith(env, "go", "build", "-o", "build/", "-ldflags", ldflags, pocketPackage)
}

func BuildRace() error {
	env[versionStringEnvVarName] += "+race"
	return sh.RunWith(env, "go", "build", "-o", "build/", "-ldflags", ldflags, "-race", pocketPackage)
}

func Install() error {
	return sh.RunWith(env, "go", "install", "-ldflags", ldflags, pocketPackage)
}

func Uninstall() error {
	return sh.Run("go", "clean", "-i", pocketPackage)
}
