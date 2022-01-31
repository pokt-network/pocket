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
	branch, _ := sh.Output("git", "branch", "--show-current")
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	_, dirty := sh.Output("git", "diff", "--quiet")

	if branch != "" {
		branch = "-" + branch
		hash = "/" + hash
	}
	if dirty != nil && dirty.Error() == "running \"git diff --quiet\" failed with exit code 1" {
		hash += "+dirty"
	}
	env[versionStringEnvVarName] = "0.0.0" + branch + hash
}

func Build() error {
	return sh.RunWith(env, "go", "build", "-o", "build/", "-ldflags", ldflags, pocketPackage)
}

func BuildRace() error {
	env["versionStringEnvVarName"] += "+race"
	return sh.RunWith(env, "go", "build", "-o", "build/", "-ldflags", ldflags, "-race", pocketPackage)
}

func Install() error {
	return sh.RunWith(env, "go", "install", "-ldflags", ldflags, pocketPackage)
}

func Uninstall() error {
	return sh.Run("go", "clean", "-i", pocketPackage)
}
