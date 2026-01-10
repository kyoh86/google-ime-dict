package syncer

import (
	"os"
	"os/user"
	"testing"
	"time"

	"github.com/kyoh86/gimedic"
)

func TestServicePushPull(t *testing.T) {
	origHostname := getHostname
	origUser := getCurrentUser
	origEnv := getEnv
	t.Cleanup(func() {
		getHostname = origHostname
		getCurrentUser = origUser
		getEnv = origEnv
	})
	getHostname = func() (string, error) { return "Host", nil }
	getCurrentUser = func() (*user.User, error) { return &user.User{Username: "User"}, nil }
	getEnv = func(string) string { return "" }

	stateHome := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateHome)

	dir := t.TempDir()
	dbPath := dir + "/user_dictionary.db"
	journalDir := dir + "/journals"

	if err := WriteStorage(dbPath, storageWithEntry("main", "k1", "v1")); err != nil {
		t.Fatalf("WriteStorage: %v", err)
	}

	service := Service{
		DBPath:          dbPath,
		JournalDir:      journalDir,
		InhibitDuration: 1 * time.Second,
	}
	journalPath, err := service.ResolveJournalPath("")
	if err != nil {
		t.Fatalf("ResolveJournalPath: %v", err)
	}
	wrote, err := service.Push(journalPath)
	if err != nil {
		t.Fatalf("Push: %v", err)
	}
	if wrote == 0 {
		t.Fatalf("expected events written")
	}

	otherJournal := journalDir + "/Other.jsonl"
	if err := os.WriteFile(otherJournal, []byte(`{"op":"add","dict":"main","key":"k2","value":"v2","pos":1}`+"\n"), 0o644); err != nil {
		t.Fatalf("write other journal: %v", err)
	}
	applied, err := service.Pull([]string{otherJournal})
	if err != nil {
		t.Fatalf("Pull: %v", err)
	}
	if applied == 0 {
		t.Fatalf("expected events applied")
	}

	storage, err := LoadStorage(dbPath)
	if err != nil {
		t.Fatalf("LoadStorage: %v", err)
	}
	entries := storage.GetDictionaries()[0].GetEntries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func storageWithEntry(dictName, key, value string) *gimedic.UserDictionaryStorage {
	id := uint64(1)
	name := dictName
	entryKey := key
	entryValue := value
	pos := gimedic.UserDictionary_NOUN
	dict := &gimedic.UserDictionary{
		Id:   &id,
		Name: &name,
		Entries: []*gimedic.UserDictionary_Entry{
			{
				Key:   &entryKey,
				Value: &entryValue,
				Pos:   &pos,
			},
		},
	}
	return &gimedic.UserDictionaryStorage{
		Dictionaries: []*gimedic.UserDictionary{dict},
	}
}
