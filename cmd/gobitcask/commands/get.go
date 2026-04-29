package commands

import (
	"encoding/json"
	"fmt"

	"github.com/ashish/gobitcask/bitcask"
	"github.com/spf13/cobra"
)

func newGetCommand() *cobra.Command {
	var debug bool
	var dataDir string

	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Retrieve a value by key",
		Long: `Retrieve a value from the database by its key.
The value will be displayed in JSON format if it's a complex type.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			// Create Bitcask instance
			db, err := bitcask.New(dataDir, &debug)
			if err != nil {
				return fmt.Errorf("failed to create Bitcask instance: %w", err)
			}
			defer db.Close()

			value, err := db.Get(key)
			if err != nil {
				return fmt.Errorf("failed to get key: %w", err)
			}

			// Pretty print the value as JSON
			jsonValue, err := json.MarshalIndent(value, "", "  ")
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "%v\n", value)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", jsonValue)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode (JSON format)")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Data directory (default: system default)")

	return cmd
}
