package clear_test

import (
	"testing"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestClearCommand(t *testing.T) {
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
			name:        "clear without force flag",
			args:        []string{"clear"},
			expectedOut: "Error: this operation will delete all data. Use --force to confirm",
			expectError: true,
		},
		{
			name:        "clear with force flag",
			args:        []string{"clear", "--force"},
			expectedOut: "Successfully cleared all data",
			expectError: false,
		},
		{
			name:        "clear with extra arguments",
			args:        []string{"clear", "--force", "extra"},
			expectedOut: "Error: accepts 0 arg(s), received 1",
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

				// Verify data is actually cleared
				if tt.name == "clear with force flag" {
					listOutput, err := helpers.ExecuteCommand(t, rootCmd, "list")
					assert.NoError(t, err)
					helpers.AssertOutputContains(t, listOutput, "No keys found")
				}
			}
		})
	}
}

func TestClearCommandWithFlags(t *testing.T) {
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
			name:        "clear with debug flag",
			args:        []string{"clear", "--debug", "--force"},
			expectedOut: "Successfully cleared all data",
			expectError: false,
		},
		{
			name:        "clear with custom data directory",
			args:        []string{"clear", "--data-dir", "testdata", "--force"},
			expectedOut: "Successfully cleared all data",
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

				// Verify data is actually cleared
				listOutput, err := helpers.ExecuteCommand(t, rootCmd, "list")
				assert.NoError(t, err)
				helpers.AssertOutputContains(t, listOutput, "No keys found")
			}
		})
	}
}
