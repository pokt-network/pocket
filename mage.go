//go:build mage

package main

import (
	"github.com/magefile/mage/sh"
	"os"
)

const pocketPackage = "github.com/pokt-network/pocket/cmd/pocket"
const ldflag = "-X main.version=$POCKET_VERSION_STRING"

func init() {
	os.Setenv("GO111MODULE", "on")
}

func Build() error {
	env := map[string]string{"POCKET_VERSION_STRING": "1.0.0-alpha"}
	return sh.RunWith(env, "go", "build", "-ldflags", ldflag, pocketPackage)
}

func BuildRace() error {
	env := map[string]string{"POCKET_VERSION_STRING": "1.0.0-alpha+race"}
	return sh.RunWith(env, "go", "build", "-ldflags", ldflag, "-race", pocketPackage)
}

func Install() error {
	env := map[string]string{"POCKET_VERSION_STRING": "1.0.0-alpha"}
	return sh.RunWith(env, "go", "install", "-ldflags", ldflag, pocketPackage)
}

func Uninstall() error {
	return sh.Run("go", "clean", "-i", pocketPackage)
}
