package helpers

import (
	"bytes"
	"os"
	"testing"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands"
	"github.com/stretchr/testify/require"
)

// TestEnv holds an isolated test environment with its own temporary data directory.
// Each test should create its own TestEnv to prevent state leakage between tests.
type TestEnv struct {
	DataDir string
	t       *testing.T
}

// New creates a new isolated test environment with a temporary data directory.
// The directory is automatically removed when the test completes.
func New(t *testing.T) *TestEnv {
	t.Helper()
	dir, err := os.MkdirTemp("", "gobitcask-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	return &TestEnv{DataDir: dir, t: t}
}

// Run executes a gobitcask CLI command in the test environment's isolated data
// directory. It injects --data-dir automatically unless already present in args.
// Returns combined stdout+stderr output and the command error.
func (e *TestEnv) Run(args ...string) (string, error) {
	e.t.Helper()
	args = e.injectDataDir(args)
	cmd := commands.NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

// injectDataDir inserts --data-dir <e.DataDir> after the subcommand name unless
// --data-dir is already present somewhere in args.
func (e *TestEnv) injectDataDir(args []string) []string {
	if len(args) == 0 {
		return args
	}
	for _, a := range args {
		if a == "--data-dir" {
			return args
		}
	}
	// Insert after the subcommand name (args[0]).
	result := make([]string, 0, len(args)+2)
	result = append(result, args[0])
	result = append(result, "--data-dir", e.DataDir)
	result = append(result, args[1:]...)
	return result
}
