package bitcask_test

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ashish/gobitcask/bitcask"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestDB opens a fresh Bitcask in a temp dir. debug selects JSON format.
func newTestDB(t *testing.T, debug bool) *bitcask.Bitcask {
	t.Helper()
	dir, err := os.MkdirTemp("", "gobitcask-unit-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	db, err := bitcask.New(dir, &debug)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

// newTestDBInDir opens a Bitcask in a specific directory (for persistence tests).
func newTestDBInDir(t *testing.T, dir string, debug bool) *bitcask.Bitcask {
	t.Helper()
	db, err := bitcask.New(dir, &debug)
	require.NoError(t, err)
	return db
}

// ─── Basic CRUD ──────────────────────────────────────────────────────────────

func TestPutAndGet(t *testing.T) {
	db := newTestDB(t, false)

	require.NoError(t, db.Put("hello", "world"))

	val, err := db.Get("hello")
	require.NoError(t, err)
	assert.Equal(t, "world", val)
}

func TestPutOverwritesKey(t *testing.T) {
	db := newTestDB(t, false)

	require.NoError(t, db.Put("k", "first"))
	require.NoError(t, db.Put("k", "second"))

	val, err := db.Get("k")
	require.NoError(t, err)
	assert.Equal(t, "second", val)
}

func TestGetMissingKey(t *testing.T) {
	db := newTestDB(t, false)

	_, err := db.Get("ghost")
	assert.ErrorContains(t, err, "key not found")
}

func TestPutGetJSONObject(t *testing.T) {
	db := newTestDB(t, false)

	user := map[string]interface{}{
		"name":  "Alice",
		"age":   float64(30),
		"admin": true,
	}
	require.NoError(t, db.Put("user:1", user))

	val, err := db.Get("user:1")
	require.NoError(t, err)

	got, ok := val.(map[string]interface{})
	require.True(t, ok, "expected map")
	assert.Equal(t, "Alice", got["name"])
	assert.Equal(t, float64(30), got["age"])
	assert.Equal(t, true, got["admin"])
}

func TestPutGetSlice(t *testing.T) {
	db := newTestDB(t, false)

	tags := []interface{}{"go", "storage", "bitcask"}
	require.NoError(t, db.Put("tags", tags))

	val, err := db.Get("tags")
	require.NoError(t, err)

	got, ok := val.([]interface{})
	require.True(t, ok, "expected slice")
	assert.Len(t, got, 3)
	assert.Equal(t, "go", got[0])
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestDelete(t *testing.T) {
	db := newTestDB(t, false)

	require.NoError(t, db.Put("bye", "world"))
	require.NoError(t, db.Delete("bye"))

	_, err := db.Get("bye")
	assert.ErrorContains(t, err, "key not found")
}

func TestDeleteMissingKey(t *testing.T) {
	db := newTestDB(t, false)

	err := db.Delete("ghost")
	assert.ErrorContains(t, err, "key not found")
}

func TestDeletedKeyNotInList(t *testing.T) {
	db := newTestDB(t, false)

	require.NoError(t, db.Put("keep", "yes"))
	require.NoError(t, db.Put("remove", "yes"))
	require.NoError(t, db.Delete("remove"))

	keys := db.ListKeys()
	assert.Contains(t, keys, "keep")
	assert.NotContains(t, keys, "remove")
}

// ─── ListKeys ─────────────────────────────────────────────────────────────────

func TestListKeysSorted(t *testing.T) {
	db := newTestDB(t, false)

	require.NoError(t, db.Put("c", "3"))
	require.NoError(t, db.Put("a", "1"))
	require.NoError(t, db.Put("b", "2"))

	keys := db.ListKeys()
	assert.Equal(t, []string{"a", "b", "c"}, keys)
}

func TestListKeysEmpty(t *testing.T) {
	db := newTestDB(t, false)
	assert.Empty(t, db.ListKeys())
}

// ─── Clear ────────────────────────────────────────────────────────────────────

func TestClear(t *testing.T) {
	db := newTestDB(t, false)

	for i := 0; i < 5; i++ {
		require.NoError(t, db.Put(fmt.Sprintf("key:%d", i), "val"))
	}
	require.Len(t, db.ListKeys(), 5)

	require.NoError(t, db.Clear())
	assert.Empty(t, db.ListKeys())
}

func TestClearThenPut(t *testing.T) {
	db := newTestDB(t, false)

	require.NoError(t, db.Put("before", "clear"))
	require.NoError(t, db.Clear())
	require.NoError(t, db.Put("after", "clear"))

	_, err := db.Get("before")
	assert.Error(t, err)

	val, err := db.Get("after")
	require.NoError(t, err)
	assert.Equal(t, "clear", val)
}

// ─── Persistence (index rebuild) ──────────────────────────────────────────────

func TestPersistenceProtoFormat(t *testing.T) {
	dir, err := os.MkdirTemp("", "gobitcask-persist-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	debug := false

	// Write data, close.
	db1 := newTestDBInDir(t, dir, debug)
	require.NoError(t, db1.Put("name", "Bitcask"))
	require.NoError(t, db1.Put("type", "storage"))
	require.NoError(t, db1.Close())

	// Reopen and verify index was rebuilt correctly.
	db2 := newTestDBInDir(t, dir, debug)
	defer db2.Close()

	val, err := db2.Get("name")
	require.NoError(t, err)
	assert.Equal(t, "Bitcask", val)

	val, err = db2.Get("type")
	require.NoError(t, err)
	assert.Equal(t, "storage", val)
}

func TestPersistenceJSONFormat(t *testing.T) {
	dir, err := os.MkdirTemp("", "gobitcask-persist-json-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	debug := true

	db1 := newTestDBInDir(t, dir, debug)
	require.NoError(t, db1.Put("lang", "Go"))
	require.NoError(t, db1.Close())

	db2 := newTestDBInDir(t, dir, debug)
	defer db2.Close()

	val, err := db2.Get("lang")
	require.NoError(t, err)
	assert.Equal(t, "Go", val)
}

func TestPersistenceDeleteSurvivesRestart(t *testing.T) {
	dir, err := os.MkdirTemp("", "gobitcask-persist-del-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	debug := false

	db1 := newTestDBInDir(t, dir, debug)
	require.NoError(t, db1.Put("gone", "bye"))
	require.NoError(t, db1.Delete("gone"))
	require.NoError(t, db1.Close())

	db2 := newTestDBInDir(t, dir, debug)
	defer db2.Close()

	_, err = db2.Get("gone")
	assert.ErrorContains(t, err, "key not found")
	assert.NotContains(t, db2.ListKeys(), "gone")
}

// ─── Serialization formats ───────────────────────────────────────────────────

func TestProtoFormatRoundTrip(t *testing.T) {
	db := newTestDB(t, false) // proto (debug=false)

	require.NoError(t, db.Put("proto:key", "proto value"))
	val, err := db.Get("proto:key")
	require.NoError(t, err)
	assert.Equal(t, "proto value", val)
}

func TestJSONFormatRoundTrip(t *testing.T) {
	db := newTestDB(t, true) // JSON (debug=true)

	require.NoError(t, db.Put("json:key", "json value"))
	val, err := db.Get("json:key")
	require.NoError(t, err)
	assert.Equal(t, "json value", val)
}

func TestJSONFormatComplex(t *testing.T) {
	db := newTestDB(t, true)

	obj := map[string]interface{}{"x": float64(1), "y": float64(2)}
	require.NoError(t, db.Put("point", obj))

	val, err := db.Get("point")
	require.NoError(t, err)

	got := val.(map[string]interface{})
	assert.Equal(t, float64(1), got["x"])
	assert.Equal(t, float64(2), got["y"])
}

// ─── Edge cases ───────────────────────────────────────────────────────────────

func TestEmptyStringValue(t *testing.T) {
	db := newTestDB(t, false)

	require.NoError(t, db.Put("empty", ""))
	val, err := db.Get("empty")
	require.NoError(t, err)
	assert.Equal(t, "", val)
}

func TestLargeValue(t *testing.T) {
	db := newTestDB(t, false)

	// Build a large but valid UTF-8 string (repeated ASCII).
	const chunk = "abcdefghijklmnopqrstuvwxyz0123456789"
	large := ""
	for len(large) < 64*1024 {
		large += chunk
	}

	require.NoError(t, db.Put("big", large))

	val, err := db.Get("big")
	require.NoError(t, err)
	assert.Equal(t, large, val)
}

func TestManyKeys(t *testing.T) {
	db := newTestDB(t, false)
	const n = 500

	for i := 0; i < n; i++ {
		require.NoError(t, db.Put(fmt.Sprintf("key:%05d", i), fmt.Sprintf("val:%d", i)))
	}

	keys := db.ListKeys()
	assert.Len(t, keys, n)

	val, err := db.Get("key:00042")
	require.NoError(t, err)
	assert.Equal(t, "val:42", val)
}

// ─── Concurrency ──────────────────────────────────────────────────────────────

func TestConcurrentWrites(t *testing.T) {
	db := newTestDB(t, false)
	const goroutines = 20
	const keysPerGoroutine = 25

	var wg sync.WaitGroup
	errs := make(chan error, goroutines*keysPerGoroutine)

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for k := 0; k < keysPerGoroutine; k++ {
				key := fmt.Sprintf("g%d:k%d", id, k)
				if err := db.Put(key, id*1000+k); err != nil {
					errs <- err
				}
			}
		}(g)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	assert.Len(t, db.ListKeys(), goroutines*keysPerGoroutine)
}

func TestConcurrentReads(t *testing.T) {
	db := newTestDB(t, false)
	const n = 50

	for i := 0; i < n; i++ {
		require.NoError(t, db.Put(fmt.Sprintf("r:%d", i), i))
	}

	var wg sync.WaitGroup
	errs := make(chan error, n*10)

	for round := 0; round < 10; round++ {
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(key string) {
				defer wg.Done()
				_, err := db.Get(key)
				if err != nil {
					errs <- err
				}
			}(fmt.Sprintf("r:%d", i))
		}
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	db := newTestDB(t, false)

	// Seed some keys so readers don't all get not-found errors.
	for i := 0; i < 10; i++ {
		require.NoError(t, db.Put(fmt.Sprintf("seed:%d", i), i))
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Writers
	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for k := 0; ; k++ {
				select {
				case <-stop:
					return
				default:
					_ = db.Put(fmt.Sprintf("w%d:%d", id, k), k)
				}
			}
		}(w)
	}

	// Readers
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_ = db.ListKeys()
				}
			}
		}()
	}

	// Let them run briefly then stop.
	time.Sleep(100 * time.Millisecond)
	close(stop)
	wg.Wait()
}
