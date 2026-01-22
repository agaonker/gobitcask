package delete_test

import (
	"testing"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCommand(t *testing.T) {
	helpers.SetupTestEnv(t)
	rootCmd := helpers.NewTestRootCommand()

	// Create a Bitcask instance for this test
	bc := helpers.CreateTestBitcask(t)
	defer bc.Close()

	// First put some test data
	_, err := helpers.ExecuteCommand(t, rootCmd, "put", "test:key", "test value")
	assert.NoError(t, err)

	tests := []struct {
		name        string
		args        []string
		expectedOut string
		expectError bool
	}{
		{
			name:        "delete existing key",
			args:        []string{"delete", "test:key"},
			expectedOut: "Successfully deleted key: test:key",
			expectError: false,
		},
		{
			name:        "delete non-existent key",
			args:        []string{"delete", "nonexistent"},
			expectedOut: "key not found",
			expectError: true,
		},
		{
			name:        "missing key argument",
			args:        []string{"delete"},
			expectedOut: "Error: accepts 1 arg(s), received 0",
			expectError: true,
		},
		{
			name:        "too many arguments",
			args:        []string{"delete", "key", "extra"},
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

func TestDeleteCommandWithFlags(t *testing.T) {
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
			name:        "delete with debug flag",
			args:        []string{"delete", "--debug", "test:key"},
			expectedOut: "Successfully deleted key: test:key",
			expectError: false,
		},
		{
			name:        "delete with custom data directory",
			args:        []string{"delete", "--data-dir", "testdata", "test:key"},
			expectedOut: "Successfully deleted key: test:key",
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
