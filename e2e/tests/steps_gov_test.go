//go:build e2e

package e2e

import (
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/stretchr/testify/require"
)

// Ensures that the test artifact DefaultParamsOwner is imported into the CLI keybase.
func (s *rootSuite) TheUserIsAnAclOwner() {
	res, err := s.validator.RunCommand("keys",
		"import",
		test_artifacts.DefaultParamsOwner.String(),
	)
	require.NoError(s, err)
	require.Contains(s, res.Stdout, "Key imported")
	res, err = s.validator.RunCommand("keys", "get", test_artifacts.DefaultParamsOwner.PublicKey().String())
	require.NoError(s, err)
	require.Contains(s, res.Stdout, test_artifacts.DefaultParamsOwner.PublicKey().String())
}
