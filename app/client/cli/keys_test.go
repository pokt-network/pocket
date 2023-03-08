package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO(0xbigboss): Add more tests for the other commands and verify the state changes rather than just the output.
// TestKeysCommands_Create is a test for the keys create command.
func TestKeysCommands_Create(t *testing.T) {
	// boilerplate that should be in a setup function for all tests
	dir, err := os.MkdirTemp("", "keys_test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// include root command for a more robust test
	remoteCLIURL = "http://localhost:8081"
	dataDir = dir
	nonInteractive = true

	// a bit more boilerplate
	cmd := keysCreateCommands()[0]
	input := bytes.NewReader([]byte("Create --non_interactive --pwd password --hint hint"))
	expectedOutput := "New Key Created"

	// Use a buffer to capture stdout
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Use a buffer to simulate stdin
	cmd.SetIn(input)

	// Execute the command with args and flags
	err = cmd.Execute()
	require.NoError(t, err)

	// TODO(0xbigboss): Add verify the state changes rather than just the output.
	// Verify the output
	output, err := io.ReadAll(buf)
	require.NoError(t, err)
	assert.Contains(t, string(output), expectedOutput)
}
