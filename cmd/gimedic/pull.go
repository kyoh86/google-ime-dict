package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/kyoh86/gimedic"
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
		journalPaths, err := resolveJournalPaths(cmd, args)
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
		applied, err := pullOnce(dbPath, journalPaths, time.Duration(inhibitSeconds)*time.Second)
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

func pullOnce(dbPath string, journalPaths []string, inhibitDuration time.Duration) (int, error) {
	appliedTotal := 0
	for _, journalPath := range journalPaths {
		statePath, err := syncStatePath(dbPath, journalPath)
		if err != nil {
			return 0, err
		}
		state, err := loadSyncState(statePath)
		if err != nil {
			return 0, err
		}

		applied, changed, newOffset, err := applyJournal(dbPath, journalPath, state.JournalOffset)
		if err != nil {
			return 0, err
		}
		appliedTotal += applied

		if changed {
			if err := setInhibit(dbPath, inhibitDuration); err != nil {
				return 0, err
			}
		}

		storage, err := loadStorage(dbPath)
		if err != nil {
			return 0, err
		}
		state.Snapshot = snapshotFromStorage(storage)
		state.JournalOffset = newOffset
		if err := saveSyncState(statePath, state); err != nil {
			return 0, err
		}
	}
	return appliedTotal, nil
}

func applyJournal(dbPath, journalPath string, offset int64) (int, bool, int64, error) {
	file, err := os.Open(journalPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, false, offset, nil
		}
		return 0, false, offset, err
	}
	defer file.Close()

	if _, err := file.Seek(offset, 0); err != nil {
		return 0, false, offset, err
	}

	reader := bufio.NewScanner(file)
	changed := false
	applied := 0

	storage, err := loadStorage(dbPath)
	if err != nil {
		return 0, false, offset, err
	}
	for reader.Scan() {
		line := reader.Bytes()
		if len(bytesTrimSpace(line)) == 0 {
			continue
		}
		var event journalEvent
		if err := json.Unmarshal(line, &event); err != nil {
			return 0, false, offset, err
		}
		if applyEvent(storage, event) {
			changed = true
			applied++
		}
	}
	if err := reader.Err(); err != nil {
		return 0, false, offset, err
	}
	if changed {
		if err := writeStorage(dbPath, storage); err != nil {
			return 0, false, offset, err
		}
	}
	newOffset, err := journalSize(journalPath)
	if err != nil {
		return applied, changed, offset, err
	}
	return applied, changed, newOffset, nil
}

func journalSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info.Size(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	return 0, err
}

func applyEvent(storage *gimedic.UserDictionaryStorage, event journalEvent) bool {
	dict := ensureDictionary(storage, event.Dict)
	key := event.Key + "\u0000" + event.Value
	for _, entry := range dict.GetEntries() {
		if entry.GetKey()+"\u0000"+entry.GetValue() == key {
			if event.Op == "delete" {
				return deleteEntry(dict, key)
			}
			return updateEntry(entry, event)
		}
	}
	if event.Op == "delete" {
		return false
	}
	return addEntry(dict, event)
}

func ensureDictionary(storage *gimedic.UserDictionaryStorage, name string) *gimedic.UserDictionary {
	if name == "" {
		name = "default"
	}
	for _, dict := range storage.GetDictionaries() {
		if dict.GetName() == name {
			return dict
		}
	}
	newID := uniqueDictionaryID(storage)
	newDict := &gimedic.UserDictionary{
		Id:      &newID,
		Name:    &name,
		Entries: []*gimedic.UserDictionary_Entry{},
	}
	storage.Dictionaries = append(storage.Dictionaries, newDict)
	return newDict
}

func deleteEntry(dict *gimedic.UserDictionary, key string) bool {
	entries := dict.GetEntries()
	for i, entry := range entries {
		if entry.GetKey()+"\u0000"+entry.GetValue() == key {
			dict.Entries = append(entries[:i], entries[i+1:]...)
			return true
		}
	}
	return false
}

func updateEntry(entry *gimedic.UserDictionary_Entry, event journalEvent) bool {
	changed := false
	if entry.GetComment() != event.Comment {
		comment := event.Comment
		entry.Comment = &comment
		changed = true
	}
	if entry.GetLocale() != event.Locale {
		locale := event.Locale
		entry.Locale = &locale
		changed = true
	}
	if entry.GetPos() != gimedic.UserDictionary_PosType(event.Pos) {
		pos := gimedic.UserDictionary_PosType(event.Pos)
		entry.Pos = &pos
		changed = true
	}
	return changed
}

func addEntry(dict *gimedic.UserDictionary, event journalEvent) bool {
	pos := gimedic.UserDictionary_PosType(event.Pos)
	key := event.Key
	value := event.Value
	comment := event.Comment
	locale := event.Locale
	entry := &gimedic.UserDictionary_Entry{
		Key:     &key,
		Value:   &value,
		Comment: &comment,
		Locale:  &locale,
		Pos:     &pos,
	}
	dict.Entries = append(dict.Entries, entry)
	return true
}
