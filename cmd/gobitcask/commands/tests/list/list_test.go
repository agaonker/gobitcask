package list_test

import (
	"testing"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCommand(t *testing.T) {
	env := helpers.New(t)

	// Seed data shared across sub-tests.
	_, err := env.Run("put", "test:key1", "value1")
	require.NoError(t, err)
	_, err = env.Run("put", "test:key2", "value2")
	require.NoError(t, err)

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

func TestListCommandFlags(t *testing.T) {
	t.Run("debug flag", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "--debug", "test:key", "test value")
		require.NoError(t, err)

		out, err := env.Run("list", "--debug")
		assert.NoError(t, err)
		assert.Contains(t, out, "test:key")
	})

	t.Run("explicit data-dir flag", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "test:key", "test value")
		require.NoError(t, err)

		out, err := env.Run("list", "--data-dir", env.DataDir)
		assert.NoError(t, err)
		assert.Contains(t, out, "test:key")
	})
}
