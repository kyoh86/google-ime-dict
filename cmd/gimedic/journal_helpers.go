package main

import "github.com/kyoh86/gimedic"

func snapshotFromStorage(storage *gimedic.UserDictionaryStorage) snapshot {
	result := snapshot{Dictionaries: map[string]map[string]entryState{}}
	for _, dict := range storage.GetDictionaries() {
		name := dict.GetName()
		if name == "" {
			name = "default"
		}
		entries := map[string]entryState{}
		for _, entry := range dict.GetEntries() {
			state := entryStateFromProto(entry)
			entries[state.Key+"\u0000"+state.Value] = state
		}
		result.Dictionaries[name] = entries
	}
	return result
}

func entryStateFromProto(entry *gimedic.UserDictionary_Entry) entryState {
	return entryState{
		Key:     entry.GetKey(),
		Value:   entry.GetValue(),
		Comment: entry.GetComment(),
		Locale:  entry.GetLocale(),
		Pos:     int32(entry.GetPos()),
	}
}

func entryStateEqual(a, b entryState) bool {
	return a.Key == b.Key &&
		a.Value == b.Value &&
		a.Comment == b.Comment &&
		a.Locale == b.Locale &&
		a.Pos == b.Pos
}

func refreshOwnSnapshot(dbPath, journalPath string) error {
	statePath, err := syncStatePath(dbPath, journalPath)
	if err != nil {
		return err
	}
	state, err := loadSyncState(statePath)
	if err != nil {
		return err
	}
	storage, err := loadStorage(dbPath)
	if err != nil {
		return err
	}
	state.Snapshot = snapshotFromStorage(storage)
	return saveSyncState(statePath, state)
}
