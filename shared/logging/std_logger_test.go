package logging

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStdLogAgent_New(t *testing.T) {
	var writer io.Writer = bufio.NewWriter(
		bytes.NewBuffer([]byte{}),
	)

	logger := CreateStdLogger(
		LOG_LEVEL_ALL,
		GLOBAL_NAMESPACE,
		"POCKET",
		writer,
	)

	assert.NotNil(
		t,
		logger,
		"logger: could not instantiate logger",
	)

	assert.NotNil(
		t,
		logger.SetLevel,
		"logger: could not retrieve the logger ref",
	)

	assert.NotNil(
		t,
		logger.SetNamespace,
		"logger: could not retrieve the logger ref",
	)
}

func TestStdLogAgent_Debug(t *testing.T) {
	var logger Logger
	var namespace Namespace = GLOBAL_NAMESPACE
	var sentence string = "logging lorem ipsum"
	var level LogLevel = LOG_LEVEL_DEBUG
	var prefix string = "POCKET"

	// initialization
	{
		logger = CreateStdLogger(
			level,
			namespace,
			prefix,
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

		logger = CreateStdLogger(level, namespace, prefix, writer)

		logger.Debug(sentence)

		writer.Flush()

		// Use contains as log line will contain YYYY/DD/MM HH:MM:SS that we can't predict and assert against
		assert.Contains(
			t,
			string(buffer.Bytes()),
			fmt.Sprintf("| [%s][%s]: [%s] %s\n", prefix, level, namespace, sentence),
		)
	}
}

func TestStdLogAgent_Info(t *testing.T) {
	var logger Logger
	var namespace Namespace = GLOBAL_NAMESPACE
	var sentence string = "logging lorem ipsum"
	var level LogLevel = LOG_LEVEL_INFO
	var prefix string = "POCKET"

	// initialization
	{
		logger = CreateStdLogger(
			level,
			namespace,
			prefix,
			bufio.NewWriter(
				bytes.NewBuffer([]byte{}),
			),
		)

		assert.NotNil(
			t,
			logger.Info,
			"logger: Info() is nil, check if it's implemented",
		)
	}

	// Debug print correction assertion
	{
		b := make([]byte, 0)
		buffer := bytes.NewBuffer(b)
		writer := bufio.NewWriter(buffer)

		logger = CreateStdLogger(level, namespace, prefix, writer)

		logger.Info(sentence)

		writer.Flush()

		// Use contains as log line will contain YYYY/DD/MM HH:MM:SS that we can't predict and assert against
		assert.Contains(
			t,
			string(buffer.Bytes()),
			fmt.Sprintf("| [%s][%s]: [%s] %s\n", prefix, level, namespace, sentence),
		)
	}
}

func TestStdLogAgent_Error(t *testing.T) {
	var logger Logger
	var namespace Namespace = GLOBAL_NAMESPACE
	var sentence string = "logging lorem ipsum"
	var level LogLevel = LOG_LEVEL_ERROR
	var prefix string = "POCKET"

	// initialization
	{
		logger = CreateStdLogger(
			level,
			namespace,
			prefix,
			bufio.NewWriter(
				bytes.NewBuffer([]byte{}),
			),
		)

		assert.NotNil(
			t,
			logger.Error,
			"logger: Info() is nil, check if it's implemented",
		)
	}

	// Debug print correction assertion
	{
		b := make([]byte, 0)
		buffer := bytes.NewBuffer(b)
		writer := bufio.NewWriter(buffer)

		logger = CreateStdLogger(level, namespace, prefix, writer)

		logger.Error(sentence)

		writer.Flush()

		// Use contains as log line will contain YYYY/DD/MM HH:MM:SS that we can't predict and assert against
		assert.Contains(
			t,
			string(buffer.Bytes()),
			fmt.Sprintf("| [%s][%s]: [%s] %s\n", prefix, level, namespace, sentence),
		)
	}
}

func TestStdLogAgent_Warn(t *testing.T) {
	var logger Logger
	var namespace Namespace = GLOBAL_NAMESPACE
	var sentence string = "logging lorem ipsum"
	var level LogLevel = LOG_LEVEL_WARN
	var prefix string = "POCKET"

	// initialization
	{
		logger = CreateStdLogger(
			level,
			namespace,
			prefix,
			bufio.NewWriter(
				bytes.NewBuffer([]byte{}),
			),
		)

		assert.NotNil(
			t,
			logger.Warn,
			"logger: Warn() is nil, check if it's implemented",
		)
	}

	// Debug print correction assertion
	{
		b := make([]byte, 0)
		buffer := bytes.NewBuffer(b)
		writer := bufio.NewWriter(buffer)

		logger = CreateStdLogger(level, namespace, prefix, writer)

		logger.Warn(sentence)

		writer.Flush()

		// Use contains as log line will contain YYYY/DD/MM HH:MM:SS that we can't predict and assert against
		assert.Contains(
			t,
			string(buffer.Bytes()),
			fmt.Sprintf("| [%s][%s]: [%s] %s\n", prefix, level, namespace, sentence),
		)
	}
}

func TestStdLogAgent_Fatal(t *testing.T) {
	var logger Logger
	var namespace Namespace = GLOBAL_NAMESPACE
	var sentence string = "logging lorem ipsum"
	var level LogLevel = LOG_LEVEL_FATAL
	var prefix string = "POCKET"

	// initialization
	{
		logger = CreateStdLogger(
			level,
			namespace,
			prefix,
			bufio.NewWriter(
				bytes.NewBuffer([]byte{}),
			),
		)

		assert.NotNil(
			t,
			logger.Fatal,
			"logger: Warn() is nil, check if it's implemented",
		)
	}

	// Debug print correction assertion
	{
		b := make([]byte, 0)
		buffer := bytes.NewBuffer(b)
		writer := bufio.NewWriter(buffer)

		logger = CreateStdLogger(level, namespace, prefix, writer)

		logger.Fatal(sentence)

		writer.Flush()

		// Use contains as log line will contain YYYY/DD/MM HH:MM:SS that we can't predict and assert against
		assert.Contains(
			t,
			string(buffer.Bytes()),
			fmt.Sprintf("| [%s][%s]: [%s] %s\n", prefix, level, namespace, sentence),
		)
	}
}
