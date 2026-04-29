package commands

import (
	"fmt"
	"path"

	"github.com/ashish/gobitcask/bitcask"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	var debug bool
	var dataDir string

	cmd := &cobra.Command{
		Use:   "list [pattern]",
		Short: "List keys in the database",
		Long: `List keys currently stored in the database, sorted alphabetically.
An optional glob pattern can be provided to filter keys (e.g. "user:*").`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := bitcask.New(dataDir, &debug)
			if err != nil {
				return fmt.Errorf("failed to create Bitcask instance: %w", err)
			}
			defer db.Close()

			keys := db.ListKeys()

			if len(args) == 1 {
				pattern := args[0]
				var matched []string
				for _, key := range keys {
					ok, err := path.Match(pattern, key)
					if err != nil {
						return fmt.Errorf("invalid pattern %q: %w", pattern, err)
					}
					if ok {
						matched = append(matched, key)
					}
				}
				if len(matched) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No keys found matching pattern: %s\n", pattern)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Found %d keys:\n", len(matched))
					for _, key := range matched {
						fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", key)
					}
				}
				return nil
			}

			if len(keys) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No keys found")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Found %d keys:\n", len(keys))
				for _, key := range keys {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", key)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode (JSON format)")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Data directory (default: system default)")

	return cmd
}
