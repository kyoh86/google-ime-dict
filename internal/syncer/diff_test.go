package syncer

import "testing"

func TestDiffSnapshots(t *testing.T) {
	before := Snapshot{
		Dictionaries: map[string]map[string]EntryState{
			"A": {
				"k1\x00v1": {Key: "k1", Value: "v1", Comment: "old", Pos: 1},
			},
			"B": {
				"k2\x00v2": {Key: "k2", Value: "v2", Pos: 2},
			},
		},
	}
	after := Snapshot{
		Dictionaries: map[string]map[string]EntryState{
			"A": {
				"k1\x00v1": {Key: "k1", Value: "v1", Comment: "new", Pos: 1},
				"k3\x00v3": {Key: "k3", Value: "v3", Pos: 3},
			},
		},
	}
	events := DiffSnapshots(before, after)
	got := map[string]JournalEvent{}
	for _, ev := range events {
		key := ev.Op + "|" + ev.Dict + "|" + ev.Key + "|" + ev.Value
		got[key] = ev
	}
	assertEvent := func(op, dict, key, value string) {
		t.Helper()
		if _, ok := got[op+"|"+dict+"|"+key+"|"+value]; !ok {
			t.Fatalf("missing event %s %s %s %s", op, dict, key, value)
		}
	}
	assertEvent("update", "A", "k1", "v1")
	assertEvent("add", "A", "k3", "v3")
	assertEvent("delete", "B", "k2", "v2")
	if len(got) != 3 {
		t.Fatalf("unexpected events: %#v", got)
	}
}
