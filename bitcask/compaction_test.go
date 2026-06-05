package bitcask_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ashish/gobitcask/bitcask"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func countDataFiles(t *testing.T, dir string) int {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "data_*.db"))
	require.NoError(t, err)
	return len(matches)
}

func TestCompactReclaimsSpace(t *testing.T) {
	dir, err := os.MkdirTemp("", "gobitcask-compact-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	debug := false
	db, err := bitcask.New(dir, &debug)
	require.NoError(t, err)

	// Write 20 keys
	for i := 0; i < 20; i++ {
		require.NoError(t, db.Put(fmt.Sprintf("key:%d", i), fmt.Sprintf("val:%d", i)))
	}

	// Overwrite first 10
	for i := 0; i < 10; i++ {
		require.NoError(t, db.Put(fmt.Sprintf("key:%d", i), fmt.Sprintf("new:%d", i)))
	}

	// Delete 5 keys
	for i := 10; i < 15; i++ {
		require.NoError(t, db.Delete(fmt.Sprintf("key:%d", i)))
	}

	// Force multiple data files by closing and reopening
	require.NoError(t, db.Close())
	db, err = bitcask.New(dir, &debug)
	require.NoError(t, err)

	// Add more data to create second data file scenario
	for i := 20; i < 30; i++ {
		require.NoError(t, db.Put(fmt.Sprintf("key:%d", i), fmt.Sprintf("val:%d", i)))
	}

	// Count files before compaction
	filesBefore := countDataFiles(t, dir)
	require.GreaterOrEqual(t, filesBefore, 2, "need at least 2 data files for meaningful compaction test")

	// Compact
	stats, err := db.Compact()
	require.NoError(t, err)
	assert.Greater(t, stats.StaleEntries, 0)

	// Verify all live keys still readable
	for i := 0; i < 10; i++ {
		val, err := db.Get(fmt.Sprintf("key:%d", i))
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("new:%d", i), val)
	}
	for i := 10; i < 15; i++ {
		_, err := db.Get(fmt.Sprintf("key:%d", i))
		assert.ErrorContains(t, err, "key not found")
	}
	for i := 15; i < 30; i++ {
		val, err := db.Get(fmt.Sprintf("key:%d", i))
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("val:%d", i), val)
	}

	require.NoError(t, db.Close())
}

func TestCompactJSONFormat(t *testing.T) {
	dir, err := os.MkdirTemp("", "gobitcask-compact-json-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	debug := true
	db, err := bitcask.New(dir, &debug)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		require.NoError(t, db.Put(fmt.Sprintf("j:%d", i), fmt.Sprintf("v:%d", i)))
	}
	// Overwrite all
	for i := 0; i < 10; i++ {
		require.NoError(t, db.Put(fmt.Sprintf("j:%d", i), fmt.Sprintf("new:%d", i)))
	}

	require.NoError(t, db.Close())
	db, err = bitcask.New(dir, &debug)
	require.NoError(t, err)

	stats, err := db.Compact()
	require.NoError(t, err)
	assert.Equal(t, 10, stats.LiveEntries)

	for i := 0; i < 10; i++ {
		val, err := db.Get(fmt.Sprintf("j:%d", i))
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("new:%d", i), val)
	}

	require.NoError(t, db.Close())
}

func TestCompactNoImmutableFilesIsNoop(t *testing.T) {
	dir, err := os.MkdirTemp("", "gobitcask-compact-noop-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	debug := false
	db, err := bitcask.New(dir, &debug)
	require.NoError(t, err)

	require.NoError(t, db.Put("only", "key"))

	stats, err := db.Compact()
	require.NoError(t, err)
	assert.Equal(t, 0, stats.FilesCompacted)
	assert.Equal(t, 0, stats.FilesCreated)

	val, err := db.Get("only")
	require.NoError(t, err)
	assert.Equal(t, "key", val)

	require.NoError(t, db.Close())
}

func TestCompactPersistence(t *testing.T) {
	dir, err := os.MkdirTemp("", "gobitcask-compact-persist-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	debug := false

	// Phase 1: write data, close to create first immutable file.
	db1, err := bitcask.New(dir, &debug)
	require.NoError(t, err)
	for i := 0; i < 15; i++ {
		require.NoError(t, db1.Put(fmt.Sprintf("p:%d", i), i))
	}
	require.NoError(t, db1.Close())

	// Phase 2: reopen, add more, compact.
	db2, err := bitcask.New(dir, &debug)
	require.NoError(t, err)
	for i := 15; i < 20; i++ {
		require.NoError(t, db2.Put(fmt.Sprintf("p:%d", i), i))
	}
	_, err = db2.Compact()
	require.NoError(t, err)
	require.NoError(t, db2.Close())

	// Phase 3: reopen — index must rebuild from compacted + active files.
	db3, err := bitcask.New(dir, &debug)
	require.NoError(t, err)
	defer db3.Close()

	keys := db3.ListKeys()
	assert.Len(t, keys, 20)

	for i := 0; i < 20; i++ {
		val, err := db3.Get(fmt.Sprintf("p:%d", i))
		require.NoError(t, err)
		assert.Equal(t, float64(i), val) // JSON numbers → float64
	}
}

func TestStartupCleansTmpFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "gobitcask-tmp-cleanup-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	debug := false
	db, err := bitcask.New(dir, &debug)
	require.NoError(t, err)
	require.NoError(t, db.Put("k", "v"))
	require.NoError(t, db.Close())

	// Simulate crashed compaction: leave a .tmp file.
	tmpFile := filepath.Join(dir, "data_999.db.tmp")
	require.NoError(t, os.WriteFile(tmpFile, []byte("garbage"), 0644))

	// Reopen — should clean up the tmp file.
	db2, err := bitcask.New(dir, &debug)
	require.NoError(t, err)
	defer db2.Close()

	_, err = os.Stat(tmpFile)
	assert.True(t, os.IsNotExist(err), ".tmp file should have been removed on startup")
}
