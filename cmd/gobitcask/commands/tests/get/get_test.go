package get_test

import (
	"testing"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestGetCommand(t *testing.T) {
	helpers.SetupTestEnv(t)
	rootCmd := helpers.NewTestRootCommand()

	// Create a Bitcask instance for this test
	bc := helpers.CreateTestBitcask(t)
	defer bc.Close()

	// First put some test data
	_, err := helpers.ExecuteCommand(t, rootCmd, "put", "test:key", "test value")
	assert.NoError(t, err)
	_, err = helpers.ExecuteCommand(t, rootCmd, "put", "test:json", `{"name": "test", "value": 123}`)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		args        []string
		expectedOut string
		expectError bool
	}{
		{
			name:        "get existing string",
			args:        []string{"get", "test:key"},
			expectedOut: "test value",
			expectError: false,
		},
		{
			name:        "get existing json",
			args:        []string{"get", "test:json"},
			expectedOut: `{"name": "test", "value": 123}`,
			expectError: false,
		},
		{
			name:        "get non-existent key",
			args:        []string{"get", "nonexistent"},
			expectedOut: "key not found",
			expectError: true,
		},
		{
			name:        "missing key argument",
			args:        []string{"get"},
			expectedOut: "Error: accepts 1 arg(s), received 0",
			expectError: true,
		},
		{
			name:        "too many arguments",
			args:        []string{"get", "key", "extra"},
			expectedOut: "Error: accepts 1 arg(s), received 2",
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

func TestGetCommandWithFlags(t *testing.T) {
	helpers.SetupTestEnv(t)
	rootCmd := helpers.NewTestRootCommand()

	// Create a Bitcask instance for this test
	bc := helpers.CreateTestBitcask(t)
	defer bc.Close()

	// First put some test data with debug mode
	_, err := helpers.ExecuteCommand(t, rootCmd, "put", "--debug", "test:key", "test value")
	assert.NoError(t, err)

	tests := []struct {
		name        string
		args        []string
		expectedOut string
		expectError bool
	}{
		{
			name:        "get with debug flag",
			args:        []string{"get", "--debug", "test:key"},
			expectedOut: "test value",
			expectError: false,
		},
		{
			name:        "get with custom data directory",
			args:        []string{"get", "--data-dir", "testdata", "test:key"},
			expectedOut: "test value",
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
