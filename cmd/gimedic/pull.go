package main

import (
	"errors"
	"time"

	"github.com/apex/log"
	"github.com/kyoh86/gimedic/internal/syncer"
	"github.com/spf13/cobra"
)

var pullCommand = &cobra.Command{
	Use:   "pull [journal.jsonl...]",
	Short: "Apply shared journal entries to local dictionary",
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
		if len(journalPaths) == 0 {
			return errors.New("no journal files found")
		}
		inhibitSeconds, err := cmd.Flags().GetInt("inhibit-seconds")
		if err != nil {
			return err
		}
		service.InhibitDuration = time.Duration(inhibitSeconds) * time.Second
		applied, err := service.Pull(journalPaths)
		if err != nil {
			return err
		}
		if applied > 0 {
			log.Infof("pull: applied %d events", applied)
		}
		return nil
	},
}

func init() {
	pullCommand.Flags().String("path", "", "Local user_dictionary.db path (overrides auto-detect)")
	pullCommand.Flags().String("journal-dir", "", "Directory for journal files (overrides default)")
	pullCommand.Flags().Int("inhibit-seconds", 2, "Seconds to inhibit push after applying changes")
	facadeCommand.AddCommand(pullCommand)
}
