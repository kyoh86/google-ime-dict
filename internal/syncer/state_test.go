package syncer

import (
	"os"
	"testing"
	"time"

	"github.com/kyoh86/gimedic"
	"google.golang.org/protobuf/proto"
)

func TestInhibitState(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dbPath := "/tmp/dummy.db"
	if err := SetInhibit(dbPath, 2*time.Second); err != nil {
		t.Fatalf("SetInhibit: %v", err)
	}
	inhibited, err := ShouldInhibit(dbPath)
	if err != nil {
		t.Fatalf("ShouldInhibit: %v", err)
	}
	if !inhibited {
		t.Fatal("expected inhibit to be true")
	}
}

func TestRefreshOwnSnapshotCreatesState(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := t.TempDir()
	dbPath := dir + "/user_dictionary.db"
	journalPath := dir + "/self.jsonl"
	storage := &gimedic.UserDictionaryStorage{}
	raw, err := proto.Marshal(storage)
	if err != nil {
		t.Fatalf("marshal storage: %v", err)
	}
	if err := os.WriteFile(dbPath, raw, 0o644); err != nil {
		t.Fatalf("write db: %v", err)
	}
	if err := RefreshOwnSnapshot(dbPath, journalPath); err != nil {
		t.Fatalf("RefreshOwnSnapshot: %v", err)
	}
	statePath, err := SyncStatePath(dbPath, journalPath)
	if err != nil {
		t.Fatalf("SyncStatePath: %v", err)
	}
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("state file missing: %v", err)
	}
}
