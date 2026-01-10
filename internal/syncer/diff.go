package syncer

import "sort"

func DiffSnapshots(before, after Snapshot) []JournalEvent {
	events := []JournalEvent{}
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

func newEvent(op, dict string, entry EntryState) JournalEvent {
	return JournalEvent{
		Op:      op,
		Dict:    dict,
		Key:     entry.Key,
		Value:   entry.Value,
		Pos:     entry.Pos,
		Comment: entry.Comment,
		Locale:  entry.Locale,
	}
}
