//go:build e2e

package e2e

import (
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/stretchr/testify/require"
)

// Ensures that the test artifact DefaultParamsOwner is imported into the CLI keybase.
func (s *rootSuite) TheUserIsAnAclOwner() {
	res, err := s.node.RunCommand("keys",
		"import",
		test_artifacts.DefaultParamsOwner.String(),
	)
	require.NoError(s, err)
	require.Contains(s, res.Stdout, "Key imported")
	res, err = s.node.RunCommand("keys", "get", test_artifacts.DefaultParamsOwner.Address().String())
	if err != nil {
		e2eLogger.Error().AnErr("error", err).Str("stdout", res.Stdout).Str("stderr", res.Stderr).Msgf("failed to get acl owner key")
		require.NoError(s, err)
	}
	require.Contains(s, res.Stdout, test_artifacts.DefaultParamsOwner.PublicKey().String())
}
