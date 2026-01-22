package list_test

import (
	"testing"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestListCommand(t *testing.T) {
	helpers.SetupTestEnv(t)
	rootCmd := helpers.NewTestRootCommand()

	// Create a Bitcask instance for this test
	bc := helpers.CreateTestBitcask(t)
	defer bc.Close()

	// First put some test data
	_, err := helpers.ExecuteCommand(t, rootCmd, "put", "test:key1", "value1")
	assert.NoError(t, err)
	_, err = helpers.ExecuteCommand(t, rootCmd, "put", "test:key2", "value2")
	assert.NoError(t, err)

	tests := []struct {
		name        string
		args        []string
		expectedOut string
		expectError bool
	}{
		{
			name:        "list all keys",
			args:        []string{"list"},
			expectedOut: "test:key1",
			expectError: false,
		},
		{
			name:        "list with pattern",
			args:        []string{"list", "test:*"},
			expectedOut: "test:key1",
			expectError: false,
		},
		{
			name:        "list with non-matching pattern",
			args:        []string{"list", "nonexistent:*"},
			expectedOut: "No keys found matching pattern: nonexistent:*",
			expectError: false,
		},
		{
			name:        "too many arguments",
			args:        []string{"list", "pattern", "extra"},
			expectedOut: "Error: accepts at most 1 arg(s), received 2",
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

func TestListCommandWithFlags(t *testing.T) {
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
			name:        "list with debug flag",
			args:        []string{"list", "--debug"},
			expectedOut: "test:key",
			expectError: false,
		},
		{
			name:        "list with custom data directory",
			args:        []string{"list", "--data-dir", "testdata"},
			expectedOut: "test:key",
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
