package helpers

import (
	"bytes"
	"os"
	"testing"

	"github.com/ashish/gobitcask/bitcask"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var (
	testDataDir string
	bc          *bitcask.Bitcask
)

// SetupTestEnv sets up the test environment
func SetupTestEnv(t *testing.T) {
	var err error
	testDataDir, err = os.MkdirTemp("", "gobitcask-test-*")
	assert.NoError(t, err)

	// Register cleanup
	t.Cleanup(func() {
		if bc != nil {
			bc.Close()
			bc = nil
		}
		if testDataDir != "" {
			os.RemoveAll(testDataDir)
		}
	})
}

// CreateTestBitcask creates a new Bitcask instance for testing
func CreateTestBitcask(t *testing.T) *bitcask.Bitcask {
	if bc != nil {
		bc.Close()
		bc = nil
	}

	debugMode := true
	var err error
	bc, err = bitcask.New(testDataDir, &debugMode)
	assert.NoError(t, err)
	return bc
}

// NewTestRootCommand creates a new root command for testing
func NewTestRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gobitcask",
		Short: "A simple key-value store",
	}
	return rootCmd
}

// ExecuteCommand executes a command and returns its output
func ExecuteCommand(t *testing.T, rootCmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

// AssertOutputContains checks if the output contains the expected string
func AssertOutputContains(t *testing.T, output, expected string) {
	assert.Contains(t, output, expected)
}

// AssertErrorContains checks if the error contains the expected string
func AssertErrorContains(t *testing.T, err error, expected string) {
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expected)
}
