package main

import (
	"bufio"
	"encoding/json"
	"os"
	"sort"
	"time"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

var pushCommand = &cobra.Command{
	Use:   "push [journal.jsonl]",
	Short: "Append local changes to a shared journal",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		journalPath, err := resolveJournalPath(cmd, args)
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

func pushOnce(dbPath, journalPath string) (int, error) {
	statePath, err := syncStatePath(dbPath, journalPath)
	if err != nil {
		return 0, err
	}
	state, err := loadSyncState(statePath)
	if err != nil {
		return 0, err
	}

	inhibited, err := shouldInhibit(dbPath)
	if err != nil {
		return 0, err
	}
	if inhibited {
		storage, err := loadStorage(dbPath)
		if err != nil {
			return 0, err
		}
		state.Snapshot = snapshotFromStorage(storage)
		return 0, saveSyncState(statePath, state)
	}

	storage, err := loadStorage(dbPath)
	if err != nil {
		return 0, err
	}

	current := snapshotFromStorage(storage)
	localEvents := diffSnapshots(state.Snapshot, current)
	if len(localEvents) > 0 {
		if err := appendJournalEvents(journalPath, localEvents); err != nil {
			return 0, err
		}
	}

	state.Snapshot = current
	if err := saveSyncState(statePath, state); err != nil {
		return 0, err
	}
	return len(localEvents), nil
}

func appendJournalEvents(path string, events []journalEvent) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, event := range events {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}
		if _, err := writer.Write(append(data, '\n')); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return file.Sync()
}

func diffSnapshots(before, after snapshot) []journalEvent {
	events := []journalEvent{}
	for dictName, afterEntries := range after.Dictionaries {
		beforeEntries := before.Dictionaries[dictName]
		for key, afterEntry := range afterEntries {
			if beforeEntry, ok := beforeEntries[key]; !ok {
				events = append(events, newEvent("add", dictName, afterEntry))
			} else if !entryStateEqual(beforeEntry, afterEntry) {
				events = append(events, newEvent("update", dictName, afterEntry))
			}
		}
		for key, beforeEntry := range beforeEntries {
			if _, ok := afterEntries[key]; !ok {
				events = append(events, newEvent("delete", dictName, beforeEntry))
			}
		}
	}
	for dictName, beforeEntries := range before.Dictionaries {
		if _, ok := after.Dictionaries[dictName]; ok {
			continue
		}
		for _, beforeEntry := range beforeEntries {
			events = append(events, newEvent("delete", dictName, beforeEntry))
		}
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].Dict == events[j].Dict {
			return events[i].Key < events[j].Key
		}
		return events[i].Dict < events[j].Dict
	})
	return events
}

func newEvent(op, dict string, entry entryState) journalEvent {
	return journalEvent{
		Op:      op,
		Dict:    dict,
		Key:     entry.Key,
		Value:   entry.Value,
		Pos:     entry.Pos,
		Comment: entry.Comment,
		Locale:  entry.Locale,
	}
}
