package bitcask

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// CompactionStats holds statistics from a compaction run.
type CompactionStats struct {
	FilesCompacted int
	FilesCreated   int
	LiveEntries    int
	StaleEntries   int
	Duration       time.Duration
}

// Compact merges all immutable data files, keeping only the latest value for
// each live key. The active file is never compacted.
//
// Algorithm:
//  1. Identify immutable file IDs (everything except activeFileID).
//  2. For each live key whose index entry points to an immutable file,
//     read the current value and re-write it to a new compacted file.
//  3. Remove old immutable files.
//  4. Invalidate file cache entries for removed files.
//  5. Update the in-memory index to point to the new locations.
func (bc *Bitcask) Compact() (*CompactionStats, error) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	start := time.Now()

	dataFiles, err := bc.getDataFiles()
	if err != nil {
		return nil, fmt.Errorf("compaction: failed to list data files: %w", err)
	}

	// Identify immutable files (not the active file).
	var immutableFileIDs []int
	for _, fid := range dataFiles {
		if fid != bc.activeFileID {
			immutableFileIDs = append(immutableFileIDs, fid)
		}
	}

	if len(immutableFileIDs) == 0 {
		return &CompactionStats{Duration: time.Since(start)}, nil
	}

	// Collect live keys that live in immutable files.
	type liveEntry struct {
		key   string
		entry *IndexEntry
	}
	var liveKeys []liveEntry

	immutableSet := make(map[int]bool, len(immutableFileIDs))
	for _, fid := range immutableFileIDs {
		immutableSet[fid] = true
	}

	for key, entry := range bc.index {
		if immutableSet[entry.FileID] {
			liveKeys = append(liveKeys, liveEntry{key: key, entry: entry})
		}
	}

	// Count total records in immutable files to compute stale count.
	totalRecords := 0
	for _, fid := range immutableFileIDs {
		totalRecords += bc.countRecordsInFile(fid)
	}
	staleCount := totalRecords - len(liveKeys)

	// Write live entries to new compacted files.
	// Pick file IDs higher than the current max to avoid collisions.
	maxID := bc.activeFileID
	for _, fid := range dataFiles {
		if fid > maxID {
			maxID = fid
		}
	}
	nextCompactedID := maxID + 1
	var compactedFiles []int

	var currentFile *os.File
	var currentFileID int
	var currentEntryCount int

	startNewCompactedFile := func() error {
		if currentFile != nil {
			if err := currentFile.Sync(); err != nil {
				return err
			}
			currentFile.Close()
		}
		currentFileID = nextCompactedID
		nextCompactedID++
		path := bc.getDataFilePath(currentFileID) + ".tmp"
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("compaction: failed to create compacted file: %w", err)
		}
		// Write format identifier.
		if _, err := f.Write([]byte{bc.format.GetFormatIdentifier()}); err != nil {
			f.Close()
			return err
		}
		currentFile = f
		currentEntryCount = 0
		compactedFiles = append(compactedFiles, currentFileID)
		return nil
	}

	// Only create compacted files if there are live keys to write.
	if len(liveKeys) > 0 {
		if err := startNewCompactedFile(); err != nil {
			return nil, err
		}

		// Write each live entry.
		for _, lk := range liveKeys {
			// Read current value from the old file.
			value, err := bc.readValueLocked(lk.entry)
			if err != nil {
				log.Printf("compaction: skipping key %q: %v", lk.key, err)
				continue
			}

			// Encode record.
			data, err := bc.format.EncodeRecord(lk.key, value, lk.entry.Timestamp)
			if err != nil {
				return nil, fmt.Errorf("compaction: failed to encode key %q: %w", lk.key, err)
			}

			// Rotate if needed.
			if currentEntryCount >= bc.maxEntriesPerFile {
				if err := startNewCompactedFile(); err != nil {
					return nil, err
				}
			}

			// Get position before write.
			pos, err := currentFile.Seek(0, 2) // seek to end
			if err != nil {
				return nil, err
			}

			if _, err := currentFile.Write(data); err != nil {
				return nil, fmt.Errorf("compaction: write failed: %w", err)
			}

			// Update index to point at new location.
			bc.index[lk.key] = &IndexEntry{
				FileID:    currentFileID,
				ValueSize: len(data),
				ValuePos:  pos,
				Timestamp: lk.entry.Timestamp,
			}
			currentEntryCount++
		}

		// Flush and close the last compacted file.
		if currentFile != nil {
			currentFile.Sync()
			currentFile.Close()
		}
	}

	// Atomically rename .tmp → .db
	for _, fid := range compactedFiles {
		tmpPath := bc.getDataFilePath(fid) + ".tmp"
		finalPath := bc.getDataFilePath(fid)
		if err := os.Rename(tmpPath, finalPath); err != nil {
			return nil, fmt.Errorf("compaction: rename failed: %w", err)
		}
	}

	// Remove old immutable files and evict from cache.
	for _, fid := range immutableFileIDs {
		// Close cached handle if present.
		if entry, ok := bc.readFileCache[fid]; ok {
			entry.File.Close()
			delete(bc.readFileCache, fid)
		}
		path := bc.getDataFilePath(fid)
		if err := os.Remove(path); err != nil {
			log.Printf("compaction: failed to remove old file %s: %v", path, err)
		}
	}

	stats := &CompactionStats{
		FilesCompacted: len(immutableFileIDs),
		FilesCreated:   len(compactedFiles),
		LiveEntries:    len(liveKeys),
		StaleEntries:   staleCount,
		Duration:       time.Since(start),
	}

	log.Printf("Compaction complete: %d files → %d files, %d live entries, %d stale entries removed in %v",
		stats.FilesCompacted, stats.FilesCreated, stats.LiveEntries, stats.StaleEntries, stats.Duration)

	return stats, nil
}

// readValueLocked reads the value for an index entry. Caller must hold bc.mutex.
func (bc *Bitcask) readValueLocked(entry *IndexEntry) (interface{}, error) {
	file, err := bc.getReadFileLocked(entry.FileID)
	if err != nil {
		return nil, err
	}

	filePath := bc.getDataFilePath(entry.FileID)
	format, err := bc.detectFormat(filePath)
	if err != nil {
		return nil, err
	}

	if _, err := file.Seek(entry.ValuePos, 0); err != nil {
		return nil, err
	}

	_, value, _, _, isTombstone, err := format.ReadRecord(file)
	if err != nil {
		return nil, err
	}
	if isTombstone {
		return nil, fmt.Errorf("unexpected tombstone")
	}
	return value, nil
}

// countRecordsInFile counts total records (including tombstones) in a data file.
func (bc *Bitcask) countRecordsInFile(fileID int) int {
	filePath := bc.getDataFilePath(fileID)
	file, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	format, err := bc.detectFormat(filePath)
	if err != nil {
		return 0
	}

	// Skip format identifier.
	if _, err := file.Seek(1, 0); err != nil {
		return 0
	}

	count := 0
	for {
		_, _, _, _, _, err := format.ReadRecord(file)
		if err != nil {
			break
		}
		count++
	}
	return count
}

// cleanupTmpFiles removes any .db.tmp files left by a crashed compaction.
func (bc *Bitcask) cleanupTmpFiles() error {
	entries, err := os.ReadDir(bc.dataDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".db.tmp") {
			path := bc.dataDir + "/" + e.Name()
			log.Printf("Cleaning up leftover tmp file: %s", path)
			if err := os.Remove(path); err != nil {
				log.Printf("Warning: failed to remove tmp file %s: %v", path, err)
			}
		}
	}
	return nil
}
