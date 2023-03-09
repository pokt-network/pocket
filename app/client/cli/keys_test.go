package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO(0xbigboss): Add more tests for the other commands and verify the state changes rather than just the output.
// TestKeysCommands_Create is a test for the keys create command.
func TestKeysCommands_Create(t *testing.T) {
	// TODO(0xbigboss): Centralize this boilerplate code into a TestMain function.
	// boilerplate that should be in a setup function for all tests
	dir, err := os.MkdirTemp("", "keys_test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// include root command for a more robust test
	remoteCLIURL = "http://localhost:8081"
	dataDir = dir
	nonInteractive = true
	kb, err := keybaseForCLI()
	require.NoError(t, err)
	addr, kps, err := kb.GetAll()
	require.NoError(t, err)
	require.Empty(t, addr)
	require.Empty(t, kps)
	require.NoError(t, kb.Stop())

	// a bit more boilerplate
	cmd := keysCreateCommands()[0]
	input := bytes.NewReader([]byte("Create --non_interactive --pwd password --hint hint"))

	// Use a buffer to capture stdout
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Use a buffer to simulate stdin
	cmd.SetIn(input)

	// Execute the command with args and flags
	err = cmd.Execute()
	require.NoError(t, err)

	// Verify a new key was created
	kb, err = keybaseForCLI()
	require.NoError(t, err)
	addr, kps, err = kb.GetAll()
	require.NoError(t, err)
	require.Len(t, addr, 1)
	require.Len(t, kps, 1)
}
