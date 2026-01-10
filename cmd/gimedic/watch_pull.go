package main

import (
	"time"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

var watchPullCommand = &cobra.Command{
	Use:   "watch-pull <journal.jsonl...>",
	Short: "Continuously apply shared journal entries to local dictionary",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, err := resolvePath(cmd, nil)
		if err != nil {
			return err
		}
		intervalSeconds, err := cmd.Flags().GetInt("interval-seconds")
		if err != nil {
			return err
		}
		inhibitSeconds, err := cmd.Flags().GetInt("inhibit-seconds")
		if err != nil {
			return err
		}

		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		applied, err := pullOnce(dbPath, args, time.Duration(inhibitSeconds)*time.Second)
		if err != nil {
			return err
		}
		if applied > 0 {
			log.Infof("watch-pull: applied %d events", applied)
		}
		for {
			select {
			case <-cmd.Context().Done():
				return nil
			case <-ticker.C:
				applied, err := pullOnce(dbPath, args, time.Duration(inhibitSeconds)*time.Second)
				if err != nil {
					return err
				}
				if applied > 0 {
					log.Infof("watch-pull: applied %d events", applied)
				}
			}
		}
	},
}

func init() {
	watchPullCommand.Flags().String("path", "", "Local user_dictionary.db path (overrides auto-detect)")
	watchPullCommand.Flags().Int("interval-seconds", 5, "Polling interval in seconds")
	watchPullCommand.Flags().Int("inhibit-seconds", 2, "Seconds to inhibit push after applying changes")
	facadeCommand.AddCommand(watchPullCommand)
}
