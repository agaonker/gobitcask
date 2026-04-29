package delete_test

import (
	"testing"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteCommand(t *testing.T) {
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
			env := helpers.New(t)

			if tt.name == "delete existing key" {
				_, err := env.Run("put", "test:key", "test value")
				require.NoError(t, err)
			}

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

func TestDeleteVerifiesKeyGone(t *testing.T) {
	env := helpers.New(t)
	_, err := env.Run("put", "mykey", "myval")
	require.NoError(t, err)

	_, err = env.Run("delete", "mykey")
	require.NoError(t, err)

	out, err := env.Run("get", "mykey")
	assert.Error(t, err)
	assert.Contains(t, out, "key not found")
}

func TestDeleteCommandFlags(t *testing.T) {
	t.Run("debug flag", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "--debug", "test:key", "test value")
		require.NoError(t, err)

		out, err := env.Run("delete", "--debug", "test:key")
		assert.NoError(t, err)
		assert.Contains(t, out, "Successfully deleted key: test:key")
	})

	t.Run("explicit data-dir flag", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "test:key", "test value")
		require.NoError(t, err)

		out, err := env.Run("delete", "--data-dir", env.DataDir, "test:key")
		assert.NoError(t, err)
		assert.Contains(t, out, "Successfully deleted key: test:key")
	})
}
