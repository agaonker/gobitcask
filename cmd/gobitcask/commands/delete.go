package commands

import (
	"fmt"

	"github.com/ashish/gobitcask/bitcask"
	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	var debug bool
	var dataDir string

	cmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a key from the database",
		Long: `Delete a key and its associated value from the database.
The operation is permanent and cannot be undone.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			// Create Bitcask instance
			db, err := bitcask.New(dataDir, &debug)
			if err != nil {
				return fmt.Errorf("failed to create Bitcask instance: %w", err)
			}
			defer db.Close()

			if err := db.Delete(key); err != nil {
				return fmt.Errorf("failed to delete key: %w", err)
			}

			fmt.Printf("Successfully deleted key: %s\n", key)
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode (JSON format)")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Data directory (default: system default)")

	return cmd
}
