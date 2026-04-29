package get_test

import (
	"testing"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCommand(t *testing.T) {
	env := helpers.New(t)

	// Seed data shared across sub-tests.
	_, err := env.Run("put", "test:key", "test value")
	require.NoError(t, err)
	_, err = env.Run("put", "test:json", `{"name": "test", "value": 123}`)
	require.NoError(t, err)

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
		},
		{
			name:        "get existing json",
			args:        []string{"get", "test:json"},
			expectedOut: `"name"`,
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
			output, err := env.Run(tt.args...)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Contains(t, output, tt.expectedOut)
		})
	}
}

func TestGetCommandFlags(t *testing.T) {
	t.Run("debug flag round-trip", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "--debug", "test:key", "test value")
		require.NoError(t, err)

		out, err := env.Run("get", "--debug", "test:key")
		assert.NoError(t, err)
		assert.Contains(t, out, "test value")
	})

	t.Run("explicit data-dir flag", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "test:key", "test value")
		require.NoError(t, err)

		out, err := env.Run("get", "--data-dir", env.DataDir, "test:key")
		assert.NoError(t, err)
		assert.Contains(t, out, "test value")
	})
}
