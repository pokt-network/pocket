package telemetry

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/assert"
)

func TestStdLogAgent_New(t *testing.T) {
	logger := NewLogger(
		modules.LOG_LEVEL_ALL,
		bufio.NewWriter(
			bytes.NewBuffer([]byte{}),
		),
	)

	assert.NotNil(
		t,
		logger,
		"logger: could not instantiate logger",
	)

	assert.NotNil(
		t,
		logger.SetOutput,
		"logger: could not retrieve the logger ref",
	)
}

func TestStdLogAgent_Debug(t *testing.T) {
	var logger *StdLogAgent
	var namespace string = "telemetry"
	var sentence string = "logging lorem ipsum"
	var level string = "[DEBUG]"
	var prefix string = "[pocket]"

	// initialization
	{
		logger = NewLogger(
			modules.LOG_LEVEL_DEBUG,
			bufio.NewWriter(
				bytes.NewBuffer([]byte{}),
			),
		)

		assert.NotNil(
			t,
			logger.Debug,
			"logger: Debug() is nil, check if it's implemented",
		)
	}

	// Debug print correction assertion
	{
		b := make([]byte, 0)
		buffer := bytes.NewBuffer(b)
		writer := bufio.NewWriter(buffer)

		logger = NewLogger(modules.LOG_LEVEL_DEBUG, writer)

		logger.Debug(namespace, sentence)

		writer.Flush()

		assert.Equal(
			t,
			fmt.Sprintf("%s%s: %s %s\n", prefix, level, namespace, sentence),
			string(buffer.Bytes()),
		)
	}

}

func TestStdLogAgent_Log(t *testing.T) {
	var logger *StdLogAgent
	var namespace string = "p2p"
	var sentence string = "logging lorem ipsum"
	var level string = "[LOG]"
	var prefix string = "[pocket]"

	// initialization
	{
		logger = NewLogger(
			modules.LOG_LEVEL_ALL,
			bufio.NewWriter(
				bytes.NewBuffer([]byte{}),
			),
		)

		assert.NotNil(
			t,
			logger.Log,
			"logger: Log() is nil, check if it's implemented",
		)
	}

	// Debug print correction assertion
	{
		b := make([]byte, 0)
		buffer := bytes.NewBuffer(b)
		writer := bufio.NewWriter(buffer)

		logger = NewLogger(modules.LOG_LEVEL_ALL, writer)

		logger.Log(namespace, sentence)

		writer.Flush()

		assert.Equal(
			t,
			fmt.Sprintf("%s%s: %s %s\n", prefix, level, namespace, sentence),
			string(buffer.Bytes()),
		)
	}
}

func TestStdLogAgent_Info(t *testing.T) {
	var logger *StdLogAgent
	var namespace = "utils"
	var sentence string = "logging lorem ipsum"
	var level string = "[INFO]"
	var prefix string = "[pocket]"

	// initialization
	{
		logger = NewLogger(
			modules.LOG_LEVEL_INFO,
			bufio.NewWriter(
				bytes.NewBuffer([]byte{}),
			),
		)

		assert.NotNil(
			t,
			logger.Info,
			"logger: INFO() is nil, check if it's implemented",
		)
	}

	// Debug print correction assertion
	{
		b := make([]byte, 0)
		buffer := bytes.NewBuffer(b)
		writer := bufio.NewWriter(buffer)

		logger = NewLogger(modules.LOG_LEVEL_INFO, writer)

		logger.Info(namespace, sentence)

		writer.Flush()

		assert.Equal(
			t,
			fmt.Sprintf("%s%s: %s %s\n", prefix, level, namespace, sentence),
			string(buffer.Bytes()),
		)
	}
}

func TestStdLogAgent_Error(t *testing.T) {
	var logger *StdLogAgent
	var namespace string = "consensus"
	var sentence string = "logging lorem ipsum"
	var level string = "[ERROR]"
	var prefix string = "[pocket]"

	// initialization
	{
		logger = NewLogger(
			modules.LOG_LEVEL_ERROR,
			bufio.NewWriter(
				bytes.NewBuffer([]byte{}),
			),
		)

		assert.NotNil(
			t,
			logger.Error,
			"logger: Error() is nil, check if it's implemented",
		)
	}

	// Debug print correction assertion
	{
		b := make([]byte, 0)
		buffer := bytes.NewBuffer(b)
		writer := bufio.NewWriter(buffer)

		logger = NewLogger(modules.LOG_LEVEL_ERROR, writer)

		logger.Error(namespace, sentence)

		writer.Flush()

		assert.Equal(
			t,
			fmt.Sprintf("%s%s: %s %s\n", prefix, level, namespace, sentence),
			string(buffer.Bytes()),
		)
	}
}
