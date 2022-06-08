package types

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_New(t *testing.T) {
	logger := NewLogger(
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

func TestLogger_Debug(t *testing.T) {
	var logger Logger
	var sentence string = "logging lorem ipsum"
	var level string = "[DEBUG]"
	var prefix string = "[pocket]"

	// initialization
	{
		logger = NewLogger(
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

		logger = NewLogger(writer)

		logger.Debug(sentence)

		writer.Flush()

		assert.Equal(
			t,
			fmt.Sprintf("%s%s: %s\n", prefix, level, sentence),
			string(buffer.Bytes()),
		)
	}

}

func TestLogger_Log(t *testing.T) {
	var logger Logger
	var sentence string = "logging lorem ipsum"
	var level string = "[LOG]"
	var prefix string = "[pocket]"

	// initialization
	{
		logger = NewLogger(
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

		logger = NewLogger(writer)

		logger.Log(sentence)

		writer.Flush()

		assert.Equal(
			t,
			fmt.Sprintf("%s%s: %s\n", prefix, level, sentence),
			string(buffer.Bytes()),
		)
	}
}

func TestLogger_Info(t *testing.T) {
	var logger Logger
	var sentence string = "logging lorem ipsum"
	var level string = "[INFO]"
	var prefix string = "[pocket]"

	// initialization
	{
		logger = NewLogger(
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

		logger = NewLogger(writer)

		logger.Info(sentence)

		writer.Flush()

		assert.Equal(
			t,
			fmt.Sprintf("%s%s: %s\n", prefix, level, sentence),
			string(buffer.Bytes()),
		)
	}
}

func TestLogger_Error(t *testing.T) {
	var logger Logger
	var sentence string = "logging lorem ipsum"
	var level string = "[ERROR]"
	var prefix string = "[pocket]"

	// initialization
	{
		logger = NewLogger(
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

		logger = NewLogger(writer)

		logger.Error(sentence)

		writer.Flush()

		assert.Equal(
			t,
			fmt.Sprintf("%s%s: %s\n", prefix, level, sentence),
			string(buffer.Bytes()),
		)
	}
}
