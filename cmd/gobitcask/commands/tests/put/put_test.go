package put_test

import (
	"testing"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestPutCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedOut string
		expectError bool
	}{
		{
			name:        "store simple string",
			args:        []string{"put", "test:key", "test value"},
			expectedOut: "Successfully stored key: test:key",
		},
		{
			name:        "store json object",
			args:        []string{"put", "test:json", `{"name": "test", "value": 123}`},
			expectedOut: "Successfully stored key: test:json",
		},
		{
			name:        "missing arguments",
			args:        []string{"put"},
			expectedOut: "Error: accepts 2 arg(s), received 0",
			expectError: true,
		},
		{
			name:        "too many arguments",
			args:        []string{"put", "key", "value", "extra"},
			expectedOut: "Error: accepts 2 arg(s), received 3",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := helpers.New(t)
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

func TestPutCommandFlags(t *testing.T) {
	t.Run("debug flag uses JSON format", func(t *testing.T) {
		env := helpers.New(t)
		out, err := env.Run("put", "--debug", "test:key", "test value")
		assert.NoError(t, err)
		assert.Contains(t, out, "Successfully stored key: test:key")
	})

	t.Run("explicit data-dir flag", func(t *testing.T) {
		env := helpers.New(t)
		out, err := env.Run("put", "--data-dir", env.DataDir, "test:key", "test value")
		assert.NoError(t, err)
		assert.Contains(t, out, "Successfully stored key: test:key")
	})
}
