//go:build e2e

package e2e

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/json"
)

func (s *rootSuite) TheUserSubmitsAMajorProtocolUpgrade() {
	res, err := s.node.RunCommand("query", "upgrade")
	require.NoError(s, err)

	var qur rpc.QueryUpgradeResponse
	// TECHDEBT: cli outputs debug logs so scan for our answer
	for _, line := range strings.Split(res.Stdout, "\n") {
		// parse into QueryUpgradeResponse
		err = json.Unmarshal([]byte(line), &qur)
		if err == nil && qur.Version != "" {
			break
		}
	}
	require.NoError(s, err)

	// submit a major protocol upgrade
	newVersion, err := semver.Parse(qur.Version)
	require.NoError(s, err)
	newVersion.Major++
	res, err = s.node.RunCommand("gov", "upgrade", test_artifacts.DefaultParams().AclOwner, newVersion.String(), fmt.Sprint(qur.Height+1))
	require.NoError(s, err)

	// TECHBDEBT: cli outputs debug logs last non-blank line is our answer
	var lines = strings.Split(res.Stdout, "\n")
	var answer string
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] != "" {
			answer = lines[i]
			break
		}
	}
	// ensure it is a valid sha256 hash
	require.Regexp(s, "^([a-f0-9]{64})$", answer, "invalid tx hash")
	s.pendingTxs = append(s.pendingTxs, answer)
	s.node.result = res
}

func (s *rootSuite) TheSystemReachesTheUpgradeHeight() {
	for {
		res, err := s.node.RunCommand("query", "upgrade")
		require.NoError(s, err)

		// parse into QueryUpgradeResponse
		var qur rpc.QueryUpgradeResponse
		err = json.Unmarshal([]byte(res.Stdout), &qur)
		require.NoError(s, err)

		if qur.Height > 0 {
			break
		}
	}
}

func (s *rootSuite) TheUserShouldBeAbleToSeeTheNewVersion() {
	panic("PENDING")
}
