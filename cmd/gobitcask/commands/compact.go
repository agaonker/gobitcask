package commands

import (
	"fmt"

	"github.com/ashish/gobitcask/bitcask"
	"github.com/spf13/cobra"
)

func newCompactCommand() *cobra.Command {
	var debug bool
	var dataDir string

	cmd := &cobra.Command{
		Use:   "compact",
		Short: "Compact data files by removing stale entries",
		Long: `Merges immutable data files, keeping only the latest value for each live key.
Reclaims disk space from overwrites and deletions.
The active file is not compacted.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := bitcask.New(dataDir, &debug)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer func() {
				if cerr := db.Close(); cerr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: failed to close db: %v\n", cerr)
				}
			}()

			stats, err := db.Compact()
			if err != nil {
				return fmt.Errorf("compaction failed: %w", err)
			}

			if stats.FilesCompacted == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Nothing to compact — only one active file.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Compaction complete:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Files compacted: %d → %d\n", stats.FilesCompacted, stats.FilesCreated)
			fmt.Fprintf(cmd.OutOrStdout(), "  Live entries:    %d\n", stats.LiveEntries)
			fmt.Fprintf(cmd.OutOrStdout(), "  Stale entries:   %d removed\n", stats.StaleEntries)
			fmt.Fprintf(cmd.OutOrStdout(), "  Duration:        %v\n", stats.Duration)
			return nil
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode (JSON format)")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Data directory (default: system default)")

	return cmd
}
