package bitcask

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ashish/gobitcask/config"
	"github.com/ashish/gobitcask/formats"
)

// IndexEntry represents an entry in the in-memory index
type IndexEntry struct {
	FileID    int   `json:"file_id"`
	ValueSize int   `json:"value_size"`
	ValuePos  int64 `json:"value_pos"`
	Timestamp int64 `json:"timestamp"`
}

// Bitcask represents a log-structured hash table for fast key/value data storage
type Bitcask struct {
	dataDir   string
	debugMode bool
	format    formats.DataFormat

	// In-memory index: key -> IndexEntry
	index map[string]*IndexEntry

	// Current active file for writing
	activeFile           *os.File
	activeFileID         int
	activeFilePath       string
	activeFileEntryCount int
	activeFileLastWrite  time.Time

	// Lock for thread safety
	mutex sync.RWMutex
}

// New creates a new Bitcask instance
func New(directory string, debugMode *bool) (*Bitcask, error) {
	// Use configuration if parameters not provided
	dataDir := config.GlobalConfig.GetDataDir(directory)
	debug := config.GlobalConfig.GetDebugMode()
	if debugMode != nil {
		debug = *debugMode
	}

	// Create format based on debug mode
	var format formats.DataFormat
	if debug {
		format = &formats.JSONFormat{}
	} else {
		format = &formats.ProtoFormat{}
	}

	bc := &Bitcask{
		dataDir:   dataDir,
		debugMode: debug,
		format:    format,
		index:     make(map[string]*IndexEntry),
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize or recover from existing data
	if err := bc.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}

	return bc, nil
}

// initialize sets up the database state from existing files
func (bc *Bitcask) initialize() error {
	// Find existing data files
	dataFiles, err := bc.getDataFiles()
	if err != nil {
		return fmt.Errorf("failed to get data files: %w", err)
	}

	if len(dataFiles) == 0 {
		// No existing data files, create new one
		bc.activeFileID = 0
		return bc.createNewDataFile()
	}

	// Use the latest data file as active
	bc.activeFileID = dataFiles[len(dataFiles)-1]
	bc.activeFilePath = bc.getDataFilePath(bc.activeFileID)

	// Open active file for appending
	if err := bc.openActiveFile(); err != nil {
		return fmt.Errorf("failed to open active file: %w", err)
	}

	// Build index from all data files
	return bc.buildIndex()
}

// getDataFiles returns a sorted list of data file IDs
func (bc *Bitcask) getDataFiles() ([]int, error) {
	files, err := os.ReadDir(bc.dataDir)
	if err != nil {
		return nil, err
	}

	var fileIDs []int
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if strings.HasPrefix(name, "data_") && strings.HasSuffix(name, ".db") {
			// Extract file ID from filename
			idStr := strings.TrimPrefix(name, "data_")
			idStr = strings.TrimSuffix(idStr, ".db")

			if id, err := strconv.Atoi(idStr); err == nil {
				fileIDs = append(fileIDs, id)
			}
		}
	}

	sort.Ints(fileIDs)
	return fileIDs, nil
}

// getDataFilePath returns the path for a data file with given ID
func (bc *Bitcask) getDataFilePath(fileID int) string {
	return filepath.Join(bc.dataDir, fmt.Sprintf("data_%d.db", fileID))
}

// createNewDataFile creates a new data file for writing
func (bc *Bitcask) createNewDataFile() error {
	if bc.activeFile != nil {
		bc.activeFile.Close()
	}

	bc.activeFileID++
	bc.activeFilePath = bc.getDataFilePath(bc.activeFileID)

	// Create the file
	file, err := os.Create(bc.activeFilePath)
	if err != nil {
		return fmt.Errorf("failed to create data file: %w", err)
	}
	file.Close()

	// Open for appending
	if err := bc.openActiveFile(); err != nil {
		return err
	}

	// Write format identifier at the start of the file
	identifier := []byte{bc.format.GetFormatIdentifier()}
	if _, err := bc.activeFile.Write(identifier); err != nil {
		return fmt.Errorf("failed to write format identifier: %w", err)
	}

	if err := bc.activeFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	bc.activeFileEntryCount = 0
	bc.activeFileLastWrite = time.Now()

	log.Printf("Created new data file: %s", bc.activeFilePath)
	return nil
}

// openActiveFile opens the active file for appending
func (bc *Bitcask) openActiveFile() error {
	file, err := os.OpenFile(bc.activeFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open active file: %w", err)
	}

	bc.activeFile = file

	// Get file size for positioning
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat active file: %w", err)
	}

	log.Printf("Opened active file: %s, size: %d", bc.activeFilePath, stat.Size())
	return nil
}

// detectFormat detects the format of a data file
func (bc *Bitcask) detectFormat(filePath string) (formats.DataFormat, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read format identifier (first byte)
	identifier := make([]byte, 1)
	if _, err := file.Read(identifier); err != nil {
		if err == io.EOF {
			// Empty file, use current format
			return bc.format, nil
		}
		return nil, err
	}

	return formats.GetFormatByIdentifier(identifier[0]), nil
}

// buildIndex builds the in-memory index from all data files
func (bc *Bitcask) buildIndex() error {
	bc.index = make(map[string]*IndexEntry)

	dataFiles, err := bc.getDataFiles()
	if err != nil {
		return err
	}

	for _, fileID := range dataFiles {
		filePath := bc.getDataFilePath(fileID)
		if err := bc.buildIndexFromFile(fileID, filePath); err != nil {
			log.Printf("Warning: failed to build index from file %s: %v", filePath, err)
			continue
		}
	}

	log.Printf("Built index with %d keys", len(bc.index))
	return nil
}

// buildIndexFromFile builds index entries from a single data file
func (bc *Bitcask) buildIndexFromFile(fileID int, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Detect format
	format, err := bc.detectFormat(filePath)
	if err != nil {
		return err
	}

	// Skip format identifier
	if _, err := file.Seek(1, 0); err != nil {
		return err
	}

	position := int64(1) // Start after format identifier

	for {
		key, _, timestamp, recordSize, isTombstone, err := format.ReadRecord(file)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if isTombstone {
			// Remove from index if it's a tombstone
			delete(bc.index, key)
		} else {
			// Add to index
			bc.index[key] = &IndexEntry{
				FileID:    fileID,
				ValueSize: recordSize,
				ValuePos:  position,
				Timestamp: timestamp,
			}
		}

		position += int64(recordSize)
	}

	return nil
}

// Put stores a key-value pair
func (bc *Bitcask) Put(key string, value interface{}) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	timestamp := time.Now().UnixNano()

	// Encode the record
	data, err := bc.format.EncodeRecord(key, value, timestamp)
	if err != nil {
		return fmt.Errorf("failed to encode record: %w", err)
	}

	// Get current position in active file
	position, err := bc.activeFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("failed to get file position: %w", err)
	}

	// Write to active file
	if _, err := bc.activeFile.Write(data); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	// Sync to disk
	if err := bc.activeFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// Update index
	bc.index[key] = &IndexEntry{
		FileID:    bc.activeFileID,
		ValueSize: len(data),
		ValuePos:  position,
		Timestamp: timestamp,
	}

	bc.activeFileEntryCount++
	bc.activeFileLastWrite = time.Now()

	return nil
}

// Get retrieves a value by key
func (bc *Bitcask) Get(key string) (interface{}, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	entry, exists := bc.index[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	// Open the data file
	filePath := bc.getDataFilePath(entry.FileID)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}
	defer file.Close()

	// Detect format
	format, err := bc.detectFormat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect format: %w", err)
	}

	// Seek to the record position
	if _, err := file.Seek(entry.ValuePos, 0); err != nil {
		return nil, fmt.Errorf("failed to seek to record: %w", err)
	}

	// Read the record
	_, value, _, _, isTombstone, err := format.ReadRecord(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read record: %w", err)
	}

	if isTombstone {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// Delete marks a key as deleted by writing a tombstone
func (bc *Bitcask) Delete(key string) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	// Check if key exists
	if _, exists := bc.index[key]; !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	timestamp := time.Now().UnixNano()

	// Encode tombstone
	data, err := bc.format.EncodeTombstone(key, timestamp)
	if err != nil {
		return fmt.Errorf("failed to encode tombstone: %w", err)
	}

	// Write to active file
	if _, err := bc.activeFile.Write(data); err != nil {
		return fmt.Errorf("failed to write tombstone: %w", err)
	}

	// Sync to disk
	if err := bc.activeFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// Remove from index
	delete(bc.index, key)

	bc.activeFileEntryCount++
	bc.activeFileLastWrite = time.Now()

	return nil
}

// ListKeys returns all keys in the database
func (bc *Bitcask) ListKeys() []string {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	keys := make([]string, 0, len(bc.index))
	for key := range bc.index {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

// Close closes the database
func (bc *Bitcask) Close() error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if bc.activeFile != nil {
		return bc.activeFile.Close()
	}

	return nil
}

// Clear removes all data from the database
func (bc *Bitcask) Clear() error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	// Close active file
	if bc.activeFile != nil {
		bc.activeFile.Close()
		bc.activeFile = nil
	}

	// Remove all data files
	dataFiles, err := bc.getDataFiles()
	if err != nil {
		return err
	}

	for _, fileID := range dataFiles {
		filePath := bc.getDataFilePath(fileID)
		if err := os.Remove(filePath); err != nil {
			log.Printf("Warning: failed to remove file %s: %v", filePath, err)
		}
	}

	// Clear index
	bc.index = make(map[string]*IndexEntry)

	// Reset state
	bc.activeFileID = 0
	bc.activeFileEntryCount = 0

	// Create new data file
	return bc.createNewDataFile()
}
