package formats

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ashish/gobitcask/proto"
	protobuf "google.golang.org/protobuf/proto"
)

// Format identifiers
const (
	FormatProto byte = 0x01 // Default format
	FormatJSON  byte = 0x02 // JSON format
)

// DataFormat defines the interface for data serialization formats
type DataFormat interface {
	GetFormatIdentifier() byte
	EncodeRecord(key string, value interface{}, timestamp int64) ([]byte, error)
	DecodeRecord(data []byte) (string, interface{}, int64, error)
	EncodeTombstone(key string, timestamp int64) ([]byte, error)
	ReadRecord(reader io.Reader) (string, interface{}, int64, int, bool, error)
}

// ProtoFormat implements Protocol Buffers format
type ProtoFormat struct{}

// GetFormatIdentifier returns the format identifier byte
func (pf *ProtoFormat) GetFormatIdentifier() byte {
	return FormatProto
}

// EncodeRecord encodes a record into protobuf format
func (pf *ProtoFormat) EncodeRecord(key string, value interface{}, timestamp int64) ([]byte, error) {
	// Convert value to JSON bytes for storage
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}

	record := &proto.Record{
		Key:       key,
		Value:     valueBytes,
		Timestamp: timestamp,
		Deleted:   false,
	}

	// Serialize the protobuf message
	protoData, err := protobuf.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	// Add size prefix (4 bytes, big-endian)
	result := make([]byte, 4+len(protoData))
	binary.BigEndian.PutUint32(result[:4], uint32(len(protoData)))
	copy(result[4:], protoData)

	return result, nil
}

// DecodeRecord decodes protobuf format into a record
func (pf *ProtoFormat) DecodeRecord(data []byte) (string, interface{}, int64, error) {
	if len(data) < 4 {
		return "", nil, 0, fmt.Errorf("data too short for size prefix")
	}

	// Skip size prefix (4 bytes)
	protoData := data[4:]

	record := &proto.Record{}
	if err := protobuf.Unmarshal(protoData, record); err != nil {
		return "", nil, 0, fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	if record.Deleted {
		return "", nil, 0, fmt.Errorf("record is a tombstone")
	}

	var value interface{}
	if err := json.Unmarshal(record.Value, &value); err != nil {
		return "", nil, 0, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return record.Key, value, record.Timestamp, nil
}

// EncodeTombstone encodes a tombstone in protobuf format
func (pf *ProtoFormat) EncodeTombstone(key string, timestamp int64) ([]byte, error) {
	record := &proto.Record{
		Key:       key,
		Timestamp: timestamp,
		Deleted:   true,
	}

	// Serialize the protobuf message
	protoData, err := protobuf.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	// Add size prefix (4 bytes, big-endian)
	result := make([]byte, 4+len(protoData))
	binary.BigEndian.PutUint32(result[:4], uint32(len(protoData)))
	copy(result[4:], protoData)

	return result, nil
}

// ReadRecord reads a record from a reader in protobuf format
func (pf *ProtoFormat) ReadRecord(reader io.Reader) (string, interface{}, int64, int, bool, error) {
	// Read the size of the protobuf message (4 bytes)
	sizeBytes := make([]byte, 4)
	if _, err := io.ReadFull(reader, sizeBytes); err != nil {
		if err == io.EOF {
			return "", nil, 0, 0, false, err
		}
		return "", nil, 0, 0, false, fmt.Errorf("failed to read size: %w", err)
	}

	size := binary.BigEndian.Uint32(sizeBytes)

	// Read the protobuf message
	data := make([]byte, size)
	if _, err := io.ReadFull(reader, data); err != nil {
		return "", nil, 0, 0, false, fmt.Errorf("failed to read data: %w", err)
	}

	record := &proto.Record{}
	if err := protobuf.Unmarshal(data, record); err != nil {
		return "", nil, 0, 0, false, fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	recordSize := int(4 + size)

	if record.Deleted {
		return record.Key, nil, record.Timestamp, recordSize, true, nil
	}

	var value interface{}
	if err := json.Unmarshal(record.Value, &value); err != nil {
		return "", nil, 0, 0, false, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return record.Key, value, record.Timestamp, recordSize, false, nil
}

// JSONFormat implements JSON format for human-readable storage
type JSONFormat struct{}

// GetFormatIdentifier returns the format identifier byte
func (jf *JSONFormat) GetFormatIdentifier() byte {
	return FormatJSON
}

// JSONRecord represents a record in JSON format
type JSONRecord struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Timestamp int64       `json:"timestamp"`
	Deleted   bool        `json:"deleted,omitempty"`
}

// EncodeRecord encodes a record into JSON format
func (jf *JSONFormat) EncodeRecord(key string, value interface{}, timestamp int64) ([]byte, error) {
	record := JSONRecord{
		Key:       key,
		Value:     value,
		Timestamp: timestamp,
	}

	data, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Add newline for line-based reading
	result := append(data, '\n')
	return result, nil
}

// DecodeRecord decodes JSON format into a record
func (jf *JSONFormat) DecodeRecord(data []byte) (string, interface{}, int64, error) {
	var record JSONRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return "", nil, 0, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return record.Key, record.Value, record.Timestamp, nil
}

// EncodeTombstone encodes a tombstone in JSON format
func (jf *JSONFormat) EncodeTombstone(key string, timestamp int64) ([]byte, error) {
	record := JSONRecord{
		Key:       key,
		Value:     nil,
		Timestamp: timestamp,
		Deleted:   true,
	}

	data, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Add newline for line-based reading
	result := append(data, '\n')
	return result, nil
}

// ReadRecord reads a record from a reader in JSON format.
// Reads one byte at a time to avoid buffering past the line boundary, which
// would corrupt sequential reads on the same io.Reader.
func (jf *JSONFormat) ReadRecord(reader io.Reader) (string, interface{}, int64, int, bool, error) {
	var line []byte
	buf := make([]byte, 1)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				break
			}
			line = append(line, buf[0])
		}
		if err == io.EOF {
			if len(line) == 0 {
				return "", nil, 0, 0, false, io.EOF
			}
			break // last line with no trailing newline
		}
		if err != nil {
			return "", nil, 0, 0, false, err
		}
	}

	var record JSONRecord
	if err := json.Unmarshal(line, &record); err != nil {
		return "", nil, 0, 0, false, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	recordSize := len(line) + 1 // +1 for the newline
	return record.Key, record.Value, record.Timestamp, recordSize, record.Deleted, nil
}

// GetFormatByIdentifier returns the appropriate format based on the identifier byte
func GetFormatByIdentifier(identifier byte) DataFormat {
	switch identifier {
	case FormatProto:
		return &ProtoFormat{}
	case FormatJSON:
		return &JSONFormat{}
	default:
		// Default to protobuf format
		return &ProtoFormat{}
	}
}
