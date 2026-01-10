package syncer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/kyoh86/gimedic"
)

func AppendJournalEvents(path string, events []JournalEvent) error {
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

func ApplyJournal(dbPath, journalPath string, offset int64) (int, bool, int64, error) {
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

	storage, err := LoadStorage(dbPath)
	if err != nil {
		return 0, false, offset, err
	}
	for reader.Scan() {
		line := reader.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var event JournalEvent
		if err := json.Unmarshal(line, &event); err != nil {
			return 0, false, offset, err
		}
		if ApplyEvent(storage, event) {
			changed = true
			applied++
		}
	}
	if err := reader.Err(); err != nil {
		return 0, false, offset, err
	}
	if changed {
		if err := WriteStorage(dbPath, storage); err != nil {
			return 0, false, offset, err
		}
	}
	newOffset, err := JournalSize(journalPath)
	if err != nil {
		return applied, changed, offset, err
	}
	return applied, changed, newOffset, nil
}

func JournalSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info.Size(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	return 0, err
}

func ApplyEvent(storage *gimedic.UserDictionaryStorage, event JournalEvent) bool {
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
