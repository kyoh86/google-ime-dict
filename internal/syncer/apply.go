package syncer

import "github.com/kyoh86/gimedic"

func ensureDictionary(storage *gimedic.UserDictionaryStorage, name string) *gimedic.UserDictionary {
	if name == "" {
		name = "default"
	}
	for _, dict := range storage.GetDictionaries() {
		if dict.GetName() == name {
			return dict
		}
	}
	newID := UniqueDictionaryID(storage)
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

func updateEntry(entry *gimedic.UserDictionary_Entry, event JournalEvent) bool {
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

func addEntry(dict *gimedic.UserDictionary, event JournalEvent) bool {
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
