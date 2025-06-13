package commands

import (
	"fmt"

	"github.com/ashish/gobitcask/bitcask"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	var debug bool
	var dataDir string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all keys in the database",
		Long: `List all keys currently stored in the database.
The keys are sorted alphabetically for easy reading.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create Bitcask instance
			db, err := bitcask.New(dataDir, &debug)
			if err != nil {
				return fmt.Errorf("failed to create Bitcask instance: %w", err)
			}
			defer db.Close()

			keys := db.ListKeys()
			if len(keys) == 0 {
				fmt.Println("No keys found")
			} else {
				fmt.Printf("Found %d keys:\n", len(keys))
				for _, key := range keys {
					fmt.Printf("  %s\n", key)
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode (JSON format)")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Data directory (default: system default)")

	return cmd
}
