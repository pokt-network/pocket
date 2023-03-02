package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKeysCreateCommands is a test for the keys create command.
// This is a simple, contrived example of how to test one command.
func TestKeysCreateCommands(t *testing.T) {

	// boilerplate that should be in a setup function for all tests
	dir, err := ioutil.TempDir("", "example")
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

	// Verify the output
	output, err := ioutil.ReadAll(buf)
	require.NoError(t, err)
	assert.Contains(t, string(output), expectedOutput)

}
