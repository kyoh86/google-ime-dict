package main

import (
	"time"

	"github.com/apex/log"
	"github.com/kyoh86/gimedic/internal/syncer"
	"github.com/spf13/cobra"
)

var watchPushCommand = &cobra.Command{
	Use:   "watch-push [journal.jsonl]",
	Short: "Continuously append local changes to a shared journal",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		journalDir, err := cmd.Flags().GetString("journal-dir")
		if err != nil {
			return err
		}
		journalPath, err := syncer.ResolveJournalPath(journalDir, firstArg(args))
		if err != nil {
			return err
		}
		dbPath, err := resolvePath(cmd, nil)
		if err != nil {
			return err
		}
		intervalSeconds, err := cmd.Flags().GetInt("interval-seconds")
		if err != nil {
			return err
		}
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		wrote, err := pushOnce(dbPath, journalPath)
		if err != nil {
			return err
		}
		if wrote > 0 {
			log.Infof("watch-push: wrote %d events to %s", wrote, journalPath)
		}
		for {
			select {
			case <-cmd.Context().Done():
				return nil
			case <-ticker.C:
				wrote, err := pushOnce(dbPath, journalPath)
				if err != nil {
					return err
				}
				if wrote > 0 {
					log.Infof("watch-push: wrote %d events to %s", wrote, journalPath)
				}
			}
		}
	},
}

func init() {
	watchPushCommand.Flags().String("path", "", "Local user_dictionary.db path (overrides auto-detect)")
	watchPushCommand.Flags().String("journal-dir", "", "Directory for journal files (overrides default)")
	watchPushCommand.Flags().Int("interval-seconds", 5, "Polling interval in seconds")
	facadeCommand.AddCommand(watchPushCommand)
}
