//go:build e2e

package e2e

import "github.com/stretchr/testify/require"

func (s *rootSuite) UserHasAValidCancelUpgradeCommandWithSignerAndVersion() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) UserHasACancelUpgradeCommandForAPastVersion() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheSystemShouldCancelTheScheduledUpgrade() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheSpecifiedUpgradeIsScheduledAndNotYetActivated() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheSystemShouldRejectTheCommandAsItCannotCancelAPastUpgrade() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheSystemShouldValidateTheCommand() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheSystemShouldSuccessfullyAcceptTheCommand() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheSystemShouldReturnTheUpdatedProtocolVersion() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheSystemShouldRejectTheCommandDueToInvalidInput() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheSystemShouldRejectTheCommandDueToTooManyVersionsAhead() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheSystemShouldReturnTheSuccessfulCancellationStatus() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheUserHasAValidUpgradeProtocolCommandWithSignerHeightAndNewVersion() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheSystemShouldApplyTheProtocolUpgradeAtTheSpecifiedActivationHeight() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheUserHasAnInvalidUpgradeProtocolCommand() {
	require.Fail(s, "implement me")
}

func (s *rootSuite) TheUserHasAUpgradeProtocolCommandWithTooManyVersionsJump() {
	require.Fail(s, "implement me")
}
