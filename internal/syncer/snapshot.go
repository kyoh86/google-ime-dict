package syncer

import "github.com/kyoh86/gimedic"

func SnapshotFromStorage(storage *gimedic.UserDictionaryStorage) Snapshot {
	result := Snapshot{Dictionaries: map[string]map[string]EntryState{}}
	for _, dict := range storage.GetDictionaries() {
		name := dict.GetName()
		if name == "" {
			name = "default"
		}
		entries := map[string]EntryState{}
		for _, entry := range dict.GetEntries() {
			state := entryStateFromProto(entry)
			entries[state.Key+"\u0000"+state.Value] = state
		}
		result.Dictionaries[name] = entries
	}
	return result
}

func entryStateFromProto(entry *gimedic.UserDictionary_Entry) EntryState {
	return EntryState{
		Key:     entry.GetKey(),
		Value:   entry.GetValue(),
		Comment: entry.GetComment(),
		Locale:  entry.GetLocale(),
		Pos:     int32(entry.GetPos()),
	}
}

func entryStateEqual(a, b EntryState) bool {
	return a.Key == b.Key &&
		a.Value == b.Value &&
		a.Comment == b.Comment &&
		a.Locale == b.Locale &&
		a.Pos == b.Pos
}

func RefreshOwnSnapshot(dbPath, journalPath string) error {
	statePath, err := SyncStatePath(dbPath, journalPath)
	if err != nil {
		return err
	}
	state, err := LoadSyncState(statePath)
	if err != nil {
		return err
	}
	storage, err := LoadStorage(dbPath)
	if err != nil {
		return err
	}
	state.Snapshot = SnapshotFromStorage(storage)
	return SaveSyncState(statePath, state)
}
