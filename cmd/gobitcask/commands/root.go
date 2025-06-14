package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewRootCommand creates the root command for the CLI
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gobitcask",
		Short: "GoBitcask is a high-performance key-value store",
		Long: `GoBitcask is a high-performance key-value store implementation in Go.
It provides efficient and reliable storage with features like:
- Append-only log storage
- In-memory index for fast lookups
- Thread-safe operations
- Data persistence
- Support for complex data types`,
	}

	// Add subcommands
	rootCmd.AddCommand(newPutCommand())
	rootCmd.AddCommand(newGetCommand())
	rootCmd.AddCommand(newDeleteCommand())
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newClearCommand())

	return rootCmd
}

// (deleted; function unused)
