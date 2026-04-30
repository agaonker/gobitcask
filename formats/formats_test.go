package formats_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/ashish/gobitcask/formats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func timestamp() int64 { return time.Now().UnixNano() }

// ─── ProtoFormat ──────────────────────────────────────────────────────────────

func TestProtoFormatIdentifier(t *testing.T) {
	f := &formats.ProtoFormat{}
	assert.Equal(t, formats.FormatProto, f.GetFormatIdentifier())
}

func TestProtoEncodeDecodeString(t *testing.T) {
	f := &formats.ProtoFormat{}
	ts := timestamp()

	data, err := f.EncodeRecord("mykey", "myvalue", ts)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	key, val, gotTS, err := f.DecodeRecord(data)
	require.NoError(t, err)
	assert.Equal(t, "mykey", key)
	assert.Equal(t, "myvalue", val)
	assert.Equal(t, ts, gotTS)
}

func TestProtoEncodeDecodeObject(t *testing.T) {
	f := &formats.ProtoFormat{}
	obj := map[string]interface{}{"n": float64(42), "ok": true}

	data, err := f.EncodeRecord("obj", obj, timestamp())
	require.NoError(t, err)

	_, val, _, err := f.DecodeRecord(data)
	require.NoError(t, err)

	got, ok := val.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(42), got["n"])
	assert.Equal(t, true, got["ok"])
}

func TestProtoTombstoneEncodeDecode(t *testing.T) {
	f := &formats.ProtoFormat{}
	ts := timestamp()

	data, err := f.EncodeTombstone("dead:key", ts)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Reading a tombstone via ReadRecord should set isTombstone=true.
	r := bytes.NewReader(data)
	key, val, _, _, isTombstone, err := f.ReadRecord(r)
	require.NoError(t, err)
	assert.True(t, isTombstone)
	assert.Equal(t, "dead:key", key)
	assert.Nil(t, val)
}

func TestProtoReadRecordEOF(t *testing.T) {
	f := &formats.ProtoFormat{}
	r := bytes.NewReader([]byte{})
	_, _, _, _, _, err := f.ReadRecord(r)
	assert.ErrorIs(t, err, io.EOF)
}

func TestProtoReadRecordSequential(t *testing.T) {
	f := &formats.ProtoFormat{}
	var buf bytes.Buffer

	keys := []string{"alpha", "beta", "gamma"}
	for i, k := range keys {
		data, err := f.EncodeRecord(k, i, timestamp())
		require.NoError(t, err)
		buf.Write(data)
	}

	r := bytes.NewReader(buf.Bytes())
	for _, expected := range keys {
		key, _, _, _, isTombstone, err := f.ReadRecord(r)
		require.NoError(t, err)
		assert.False(t, isTombstone)
		assert.Equal(t, expected, key)
	}

	// Next read should be EOF.
	_, _, _, _, _, err := f.ReadRecord(r)
	assert.ErrorIs(t, err, io.EOF)
}

// ─── JSONFormat ───────────────────────────────────────────────────────────────

func TestJSONFormatIdentifier(t *testing.T) {
	f := &formats.JSONFormat{}
	assert.Equal(t, formats.FormatJSON, f.GetFormatIdentifier())
}

func TestJSONEncodeDecodeString(t *testing.T) {
	f := &formats.JSONFormat{}
	ts := timestamp()

	data, err := f.EncodeRecord("jkey", "jvalue", ts)
	require.NoError(t, err)
	// JSON lines end with a newline.
	assert.Equal(t, byte('\n'), data[len(data)-1])

	key, val, gotTS, err := f.DecodeRecord(data[:len(data)-1]) // strip newline for Decode
	require.NoError(t, err)
	assert.Equal(t, "jkey", key)
	assert.Equal(t, "jvalue", val)
	assert.Equal(t, ts, gotTS)
}

func TestJSONTombstoneEncodeDecode(t *testing.T) {
	f := &formats.JSONFormat{}

	data, err := f.EncodeTombstone("gone", timestamp())
	require.NoError(t, err)

	r := bytes.NewReader(data)
	key, val, _, _, isTombstone, err := f.ReadRecord(r)
	require.NoError(t, err)
	assert.True(t, isTombstone)
	assert.Equal(t, "gone", key)
	assert.Nil(t, val)
}

func TestJSONReadRecordEOF(t *testing.T) {
	f := &formats.JSONFormat{}
	r := bytes.NewReader([]byte{})
	_, _, _, _, _, err := f.ReadRecord(r)
	assert.ErrorIs(t, err, io.EOF)
}

func TestJSONReadRecordSequential(t *testing.T) {
	f := &formats.JSONFormat{}
	var buf bytes.Buffer

	keys := []string{"one", "two", "three"}
	for i, k := range keys {
		data, err := f.EncodeRecord(k, i, timestamp())
		require.NoError(t, err)
		buf.Write(data)
	}

	r := bytes.NewReader(buf.Bytes())
	for _, expected := range keys {
		key, _, _, _, isTombstone, err := f.ReadRecord(r)
		require.NoError(t, err)
		assert.False(t, isTombstone)
		assert.Equal(t, expected, key)
	}

	_, _, _, _, _, err := f.ReadRecord(r)
	assert.ErrorIs(t, err, io.EOF)
}

// ─── Format registry ──────────────────────────────────────────────────────────

func TestGetFormatByIdentifier(t *testing.T) {
	proto := formats.GetFormatByIdentifier(formats.FormatProto)
	assert.Equal(t, formats.FormatProto, proto.GetFormatIdentifier())

	json := formats.GetFormatByIdentifier(formats.FormatJSON)
	assert.Equal(t, formats.FormatJSON, json.GetFormatIdentifier())

	// Unknown identifier falls back to proto.
	fallback := formats.GetFormatByIdentifier(0xFF)
	assert.Equal(t, formats.FormatProto, fallback.GetFormatIdentifier())
}
