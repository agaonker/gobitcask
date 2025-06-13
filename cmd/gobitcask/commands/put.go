package commands

import (
	"encoding/json"
	"fmt"

	"github.com/ashish/gobitcask/bitcask"
	"github.com/spf13/cobra"
)

func newPutCommand() *cobra.Command {
	var debug bool
	var dataDir string

	cmd := &cobra.Command{
		Use:   "put <key> <value>",
		Short: "Store a key-value pair",
		Long: `Store a key-value pair in the database.
The value can be a simple string or a JSON object.
Example: gobitcask put user:123 '{"name": "Alice", "age": 30}'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			valueStr := args[1]

			// Create Bitcask instance
			db, err := bitcask.New(dataDir, &debug)
			if err != nil {
				return fmt.Errorf("failed to create Bitcask instance: %w", err)
			}
			defer db.Close()

			// Try to parse value as JSON, fallback to string
			var value interface{}
			if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
				value = valueStr
			}

			if err := db.Put(key, value); err != nil {
				return fmt.Errorf("failed to put key-value: %w", err)
			}

			fmt.Printf("Successfully stored key: %s\n", key)
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode (JSON format)")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Data directory (default: system default)")

	return cmd
}
