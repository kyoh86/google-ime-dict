package syncer

import (
	"os"
	"strings"
	"testing"

	"github.com/kyoh86/gimedic"
)

func TestApplyJournalAddUpdateDelete(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/user_dictionary.db"
	journalPath := dir + "/journal.jsonl"

	if err := WriteStorage(dbPath, emptyStorage()); err != nil {
		t.Fatalf("WriteStorage: %v", err)
	}
	journal := []string{
		`{"op":"add","dict":"main","key":"k1","value":"v1","pos":1,"comment":"c1"}`,
		`{"op":"update","dict":"main","key":"k1","value":"v1","pos":2,"comment":"c2"}`,
		`{"op":"delete","dict":"main","key":"k1","value":"v1","pos":2}`,
	}
	if err := os.WriteFile(journalPath, []byte(joinLines(journal)), 0o644); err != nil {
		t.Fatalf("write journal: %v", err)
	}

	applied, changed, _, err := ApplyJournal(dbPath, journalPath, 0)
	if err != nil {
		t.Fatalf("ApplyJournal: %v", err)
	}
	if applied != 3 || !changed {
		t.Fatalf("unexpected apply result: applied=%d changed=%v", applied, changed)
	}

	storage, err := LoadStorage(dbPath)
	if err != nil {
		t.Fatalf("LoadStorage: %v", err)
	}
	if len(storage.GetDictionaries()) != 1 {
		t.Fatalf("unexpected dictionaries: %d", len(storage.GetDictionaries()))
	}
	if len(storage.GetDictionaries()[0].GetEntries()) != 0 {
		t.Fatalf("expected entry deleted")
	}
}

func TestApplyJournalOffset(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/user_dictionary.db"
	journalPath := dir + "/journal.jsonl"

	if err := WriteStorage(dbPath, emptyStorage()); err != nil {
		t.Fatalf("WriteStorage: %v", err)
	}
	journal := []string{
		`{"op":"add","dict":"main","key":"k1","value":"v1","pos":1}`,
		`{"op":"add","dict":"main","key":"k2","value":"v2","pos":1}`,
	}
	content := []byte(strings.Join(journal, "\n") + "\n")
	if err := os.WriteFile(journalPath, content, 0o644); err != nil {
		t.Fatalf("write journal: %v", err)
	}
	offset := int64(len(content) - len(journal[1]) - 1)
	applied, changed, _, err := ApplyJournal(dbPath, journalPath, offset)
	if err != nil {
		t.Fatalf("ApplyJournal: %v", err)
	}
	if applied != 1 || !changed {
		t.Fatalf("unexpected apply result: applied=%d changed=%v", applied, changed)
	}
}

func emptyStorage() *gimedic.UserDictionaryStorage {
	return &gimedic.UserDictionaryStorage{}
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n") + "\n"
}
