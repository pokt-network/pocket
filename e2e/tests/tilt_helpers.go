// // go:build e2e
package e2e

import (
	"log"
	"os/exec"
)

// HACK: Dynamic scaling actors using `p1` and the `e2e test framework` is still a WIP so this is a
// functional interim solution until there's a need for a proper design.
func (s *rootSuite) syncLocalNetConfigFromHostToLocalFS() {
	if !isPackageInstalled("tilt") {
		e2eLogger.Debug().Msgf("syncLocalNetConfigFromHostToLocalFS: 'tilt' is not installed, skipping...")
		return
	}
	sedCmd := exec.Command("tilt", "trigger", "syncback_localnet_config")
	err := sedCmd.Run()
	if err != nil {
		e2eLogger.Err(err).Msgf("syncLocalNetConfigFromHostToLocalFS: failed to run command: '%s'", sedCmd.String())
		log.Fatal(err)
	}
}

func isPackageInstalled(pkg string) bool {
	_, err := exec.LookPath(pkg)
	// check error
	if err != nil {
		// the executable is not found, return false
		if execErr, ok := err.(*exec.Error); ok && execErr.Err == exec.ErrNotFound {
			return false
		}
		// another kind of error happened, let's log and exit
		log.Fatal(err)
	}
	return true
}
