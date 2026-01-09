package main

import (
	"time"

	"github.com/spf13/cobra"
)

var watchPushCommand = &cobra.Command{
	Use:   "watch-push <journal.jsonl> [from.db]",
	Short: "Continuously append local changes to a shared journal",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		journalPath := args[0]
		dbPath, err := resolvePath(cmd, args[1:])
		if err != nil {
			return err
		}
		intervalSeconds, err := cmd.Flags().GetInt("interval-seconds")
		if err != nil {
			return err
		}
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		if err := pushOnce(dbPath, journalPath); err != nil {
			return err
		}
		for {
			select {
			case <-cmd.Context().Done():
				return nil
			case <-ticker.C:
				if err := pushOnce(dbPath, journalPath); err != nil {
					return err
				}
			}
		}
	},
}

func init() {
	watchPushCommand.Flags().String("path", "", "Local user_dictionary.db path (overrides auto-detect)")
	watchPushCommand.Flags().Int("interval-seconds", 5, "Polling interval in seconds")
	facadeCommand.AddCommand(watchPushCommand)
}
