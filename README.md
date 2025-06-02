# Go Bitcask Implementation

A high-performance Go implementation of the Bitcask storage engine, providing efficient and reliable key-value storage.

## Features

- Append-only log storage
- In-memory index for fast lookups
- Thread-safe operations with sync.RWMutex
- Data persistence
- Support for complex data types with dual serialization:
  - **Protocol Buffers** (normal mode) - Compact binary format for production
  - **JSON** (debug mode) - Human-readable format for development
- Tombstone-based deletion
- File rotation and compaction
- CLI interface

## Quick Start

```bash
# Clone the repository
git clone https://github.com/ashish/gobitcask.git
cd gobitcask

# Build the project
go build ./cmd/gobitcask

# Run basic operations
./gobitcask put user:123 '{"name": "Alice", "age": 30}'
./gobitcask get user:123
./gobitcask list
```

## Installation

1. Ensure you have Go 1.21 or later installed
2. Clone the repository:
   ```bash
   git clone https://github.com/ashish/gobitcask.git
   cd gobitcask
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Build the project:
   ```bash
   go build ./cmd/gobitcask
   ```

## Development

This Go implementation follows the same architecture as the Python version but leverages Go's performance advantages:

- **Concurrency**: Uses Go's goroutines and channels for better performance
- **Memory Management**: More efficient memory usage with Go's garbage collector
- **Type Safety**: Compile-time type checking
- **Performance**: Significantly faster than the Python implementation

## Architecture

The implementation consists of several key components:

- `bitcask/`: Core storage engine
- `formats/`: Data serialization formats (Protocol Buffers & JSON)
- `config/`: Configuration management
- `proto/`: Protocol buffer definitions
- `cmd/`: Command-line interface

## Use Cases

Same as the Python implementation:

1. **High-Write Throughput Applications**
   - Log aggregation systems
   - Event sourcing systems
   - Real-time analytics data collection
   - IoT device data storage

2. **Simple Key-Value Storage Needs**
   - Session storage
   - User preferences
   - Configuration management
   - Cache persistence

## Performance

The Go implementation provides significant performance improvements over the Python version:

- **Faster I/O**: Native system calls without Python overhead
- **Better Concurrency**: Goroutines for concurrent operations
- **Lower Memory Usage**: More efficient memory management
- **Faster Serialization**: Native Protocol Buffer support

## Roadmap

Following the same roadmap as the Python implementation:

- [x] Core storage engine
- [ ] Log compaction mechanism
- [ ] REST API server
- [ ] Distributed features (future)

## License

MIT License 