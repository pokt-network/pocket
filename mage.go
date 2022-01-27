//go:build mage

package main

import (
	"github.com/magefile/mage/sh"
	"os"
)

const pocketPackage = "github.com/pokt-network/pocket/cmd/pocket"

func init() {
	os.Setenv("GO111MODULE", "on")
}

func Build() error {
	return sh.Run("go", "build", pocketPackage)
}

func BuildRace() error {
	return sh.Run("go", "build", "-race", pocketPackage)
}

func Install() error {
	return sh.Run("go", "install", pocketPackage)
}

func Uninstall() error {
	return sh.Run("go", "clean", "-i", pocketPackage)
}
