package clear_test

import (
	"testing"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClearCommand(t *testing.T) {
	t.Run("clear without force flag is rejected", func(t *testing.T) {
		env := helpers.New(t)
		out, err := env.Run("clear")
		assert.Error(t, err)
		assert.Contains(t, out, "Error: this operation will delete all data. Use --force to confirm")
	})

	t.Run("clear with force flag wipes all data", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "test:key1", "value1")
		require.NoError(t, err)
		_, err = env.Run("put", "test:key2", "value2")
		require.NoError(t, err)

		out, err := env.Run("clear", "--force")
		assert.NoError(t, err)
		assert.Contains(t, out, "Successfully cleared database")

		// Verify all keys are gone.
		listOut, err := env.Run("list")
		assert.NoError(t, err)
		assert.Contains(t, listOut, "No keys found")
	})

	t.Run("clear rejects extra arguments", func(t *testing.T) {
		env := helpers.New(t)
		out, err := env.Run("clear", "--force", "extra")
		assert.Error(t, err)
		assert.Contains(t, out, "Error: unknown command \"extra\"")
	})
}

func TestClearCommandFlags(t *testing.T) {
	t.Run("debug flag", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "--debug", "test:key", "test value")
		require.NoError(t, err)

		out, err := env.Run("clear", "--debug", "--force")
		assert.NoError(t, err)
		assert.Contains(t, out, "Successfully cleared database")

		listOut, err := env.Run("list", "--debug")
		assert.NoError(t, err)
		assert.Contains(t, listOut, "No keys found")
	})

	t.Run("explicit data-dir flag", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "test:key", "test value")
		require.NoError(t, err)

		out, err := env.Run("clear", "--data-dir", env.DataDir, "--force")
		assert.NoError(t, err)
		assert.Contains(t, out, "Successfully cleared database")
	})
}
