package compact_test

import (
	"os"
	"testing"

	"github.com/ashish/gobitcask/bitcask"
	"github.com/ashish/gobitcask/cmd/gobitcask/commands/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompactCommandNoop(t *testing.T) {
	t.Run("compact with only active file reports nothing to compact", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "k1", "v1")
		require.NoError(t, err)

		out, err := env.Run("compact")
		assert.NoError(t, err)
		assert.Contains(t, out, "Nothing to compact")
	})
}

func TestCompactCommandWithData(t *testing.T) {
	t.Run("compact after rotation reports stats", func(t *testing.T) {
		// Seed data with rotation using the library (CLI has no rotate command).
		dir, err := os.MkdirTemp("", "gobitcask-cli-compact-*")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		debug := false
		db, err := bitcask.New(dir, &debug)
		require.NoError(t, err)

		for i := 0; i < 5; i++ {
			require.NoError(t, db.Put("k", "overwrite"))
		}
		require.NoError(t, db.RotateActiveFile())
		require.NoError(t, db.Put("k", "final"))
		require.NoError(t, db.Close())

		// Now run compact via CLI.
		env := helpers.New(t)
		env.DataDir = dir
		out, err := env.Run("compact")
		assert.NoError(t, err)
		assert.Contains(t, out, "Compaction complete")
		assert.Contains(t, out, "Live entries:")
		assert.Contains(t, out, "Stale entries:")
	})
}

func TestCompactCommandRejectsArgs(t *testing.T) {
	t.Run("compact rejects extra arguments", func(t *testing.T) {
		env := helpers.New(t)
		out, err := env.Run("compact", "extra")
		assert.Error(t, err)
		assert.Contains(t, out, "Error: unknown command")
	})
}

func TestCompactCommandFlags(t *testing.T) {
	t.Run("debug flag", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "--debug", "k", "v")
		require.NoError(t, err)

		out, err := env.Run("compact", "--debug")
		assert.NoError(t, err)
		assert.Contains(t, out, "Nothing to compact")
	})

	t.Run("explicit data-dir flag", func(t *testing.T) {
		env := helpers.New(t)
		_, err := env.Run("put", "k", "v")
		require.NoError(t, err)

		out, err := env.Run("compact", "--data-dir", env.DataDir)
		assert.NoError(t, err)
		assert.Contains(t, out, "Nothing to compact")
	})
}
