package main

import (
	"errors"
	"time"

	"github.com/apex/log"
	"github.com/kyoh86/gimedic/internal/syncer"
	"github.com/spf13/cobra"
)

var watchPullCommand = &cobra.Command{
	Use:   "watch-pull [journal.jsonl...]",
	Short: "Continuously apply shared journal entries to local dictionary",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, err := resolvePath(cmd, nil)
		if err != nil {
			return err
		}
		journalDir, err := cmd.Flags().GetString("journal-dir")
		if err != nil {
			return err
		}
		service := syncer.Service{
			DBPath:     dbPath,
			JournalDir: journalDir,
		}
		journalPaths, err := service.ResolveJournalPaths(args)
		if err != nil {
			return err
		}
		if len(journalPaths) == 0 && len(args) > 0 {
			return errors.New("no journal files found")
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

		if len(journalPaths) > 0 {
			service.InhibitDuration = time.Duration(inhibitSeconds) * time.Second
			applied, err := service.Pull(journalPaths)
			if err != nil {
				return err
			}
			if applied > 0 {
				log.Infof("watch-pull: applied %d events", applied)
			}
		}
		for {
			select {
			case <-cmd.Context().Done():
				return nil
			case <-ticker.C:
				paths := journalPaths
				if len(args) == 0 {
					paths, err = service.ResolveJournalPaths(args)
					if err != nil {
						return err
					}
					if len(paths) == 0 {
						continue
					}
				}
				applied, err := service.Pull(paths)
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
	watchPullCommand.Flags().String("journal-dir", "", "Directory for journal files (overrides default)")
	watchPullCommand.Flags().Int("interval-seconds", 5, "Polling interval in seconds")
	watchPullCommand.Flags().Int("inhibit-seconds", 2, "Seconds to inhibit push after applying changes")
	facadeCommand.AddCommand(watchPullCommand)
}
