package main

import (
	"github.com/apex/log"
	"github.com/kyoh86/gimedic/internal/syncer"
	"github.com/spf13/cobra"
)

var pushCommand = &cobra.Command{
	Use:   "push [journal.jsonl]",
	Short: "Append local changes to a shared journal",
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
		wrote, err := pushOnce(dbPath, journalPath)
		if err != nil {
			return err
		}
		if wrote > 0 {
			log.Infof("push: wrote %d events to %s", wrote, journalPath)
		}
		return nil
	},
}

func init() {
	pushCommand.Flags().String("path", "", "Local user_dictionary.db path (overrides auto-detect)")
	pushCommand.Flags().String("journal-dir", "", "Directory for journal files (overrides default)")
	facadeCommand.AddCommand(pushCommand)
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

func pushOnce(dbPath, journalPath string) (int, error) {
	statePath, err := syncer.SyncStatePath(dbPath, journalPath)
	if err != nil {
		return 0, err
	}
	state, err := syncer.LoadSyncState(statePath)
	if err != nil {
		return 0, err
	}

	inhibited, err := syncer.ShouldInhibit(dbPath)
	if err != nil {
		return 0, err
	}
	if inhibited {
		return 0, nil
	}

	storage, err := syncer.LoadStorage(dbPath)
	if err != nil {
		return 0, err
	}

	current := syncer.SnapshotFromStorage(storage)
	localEvents := syncer.DiffSnapshots(state.Snapshot, current)
	if len(localEvents) > 0 {
		if err := syncer.AppendJournalEvents(journalPath, localEvents); err != nil {
			return 0, err
		}
	}

	state.Snapshot = current
	if err := syncer.SaveSyncState(statePath, state); err != nil {
		return 0, err
	}
	return len(localEvents), nil
}
