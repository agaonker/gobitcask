package put_test

import (
	"testing"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestPutCommand(t *testing.T) {
	helpers.SetupTestEnv(t)
	rootCmd := helpers.NewTestRootCommand()

	// Create a Bitcask instance for this test
	bc := helpers.CreateTestBitcask(t)
	defer bc.Close()

	tests := []struct {
		name        string
		args        []string
		expectedOut string
		expectError bool
	}{
		{
			name:        "store simple string",
			args:        []string{"put", "test:key", "test value"},
			expectedOut: "Successfully stored key: test:key",
			expectError: false,
		},
		{
			name:        "store json object",
			args:        []string{"put", "test:json", `{"name": "test", "value": 123}`},
			expectedOut: "Successfully stored key: test:json",
			expectError: false,
		},
		{
			name:        "missing arguments",
			args:        []string{"put"},
			expectedOut: "Error: accepts 2 arg(s), received 0",
			expectError: true,
		},
		{
			name:        "too many arguments",
			args:        []string{"put", "key", "value", "extra"},
			expectedOut: "Error: accepts 2 arg(s), received 3",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := helpers.ExecuteCommand(t, rootCmd, tt.args...)

			if tt.expectError {
				assert.Error(t, err)
				helpers.AssertOutputContains(t, output, tt.expectedOut)
			} else {
				assert.NoError(t, err)
				helpers.AssertOutputContains(t, output, tt.expectedOut)
			}
		})
	}
}

func TestPutCommandWithFlags(t *testing.T) {
	helpers.SetupTestEnv(t)
	rootCmd := helpers.NewTestRootCommand()

	// Create a Bitcask instance for this test
	bc := helpers.CreateTestBitcask(t)
	defer bc.Close()

	tests := []struct {
		name        string
		args        []string
		expectedOut string
		expectError bool
	}{
		{
			name:        "put with debug flag",
			args:        []string{"put", "--debug", "test:key", "test value"},
			expectedOut: "Successfully stored key: test:key",
			expectError: false,
		},
		{
			name:        "put with custom data directory",
			args:        []string{"put", "--data-dir", "testdata", "test:key", "test value"},
			expectedOut: "Successfully stored key: test:key",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := helpers.ExecuteCommand(t, rootCmd, tt.args...)

			if tt.expectError {
				assert.Error(t, err)
				helpers.AssertOutputContains(t, output, tt.expectedOut)
			} else {
				assert.NoError(t, err)
				helpers.AssertOutputContains(t, output, tt.expectedOut)
			}
		})
	}
}
