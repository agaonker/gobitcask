package commands

import (
	"fmt"

	"github.com/ashish/gobitcask/bitcask"
	"github.com/spf13/cobra"
)

func newClearCommand() *cobra.Command {
	var debug bool
	var dataDir string
	var force bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all data from the database",
		Long: `Clear all data from the database.
This operation is permanent and cannot be undone.
Use with caution!`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				return fmt.Errorf("this operation will delete all data. Use --force to confirm")
			}

			// Create Bitcask instance
			db, err := bitcask.New(dataDir, &debug)
			if err != nil {
				return fmt.Errorf("failed to create Bitcask instance: %w", err)
			}
			defer db.Close()

			if err := db.Clear(); err != nil {
				return fmt.Errorf("failed to clear database: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Successfully cleared database")
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode (JSON format)")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Data directory (default: system default)")
	cmd.Flags().BoolVar(&force, "force", false, "Force clear operation without confirmation")

	return cmd
}
