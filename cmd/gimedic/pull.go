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
		journalPaths, err := syncer.ResolveJournalPaths(journalDir, args)
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
		selfJournalPath, err := syncer.OwnJournalPath(journalDir)
		if err != nil {
			return err
		}
		applied, err := pullOnce(dbPath, journalPaths, selfJournalPath, time.Duration(inhibitSeconds)*time.Second)
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

func pullOnce(dbPath string, journalPaths []string, selfJournalPath string, inhibitDuration time.Duration) (int, error) {
	appliedTotal := 0
	for _, journalPath := range journalPaths {
		statePath, err := syncer.SyncStatePath(dbPath, journalPath)
		if err != nil {
			return 0, err
		}
		state, err := syncer.LoadSyncState(statePath)
		if err != nil {
			return 0, err
		}

		applied, changed, newOffset, err := syncer.ApplyJournal(dbPath, journalPath, state.JournalOffset)
		if err != nil {
			return 0, err
		}
		appliedTotal += applied

		if changed {
			if err := syncer.SetInhibit(dbPath, inhibitDuration); err != nil {
				return 0, err
			}
		}

		storage, err := syncer.LoadStorage(dbPath)
		if err != nil {
			return 0, err
		}
		state.Snapshot = syncer.SnapshotFromStorage(storage)
		state.JournalOffset = newOffset
		if err := syncer.SaveSyncState(statePath, state); err != nil {
			return 0, err
		}
	}
	if err := syncer.RefreshOwnSnapshot(dbPath, selfJournalPath); err != nil {
		return appliedTotal, err
	}
	return appliedTotal, nil
}
