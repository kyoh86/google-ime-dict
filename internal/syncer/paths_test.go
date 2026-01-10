package syncer

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestJournalIdentitySanitizes(t *testing.T) {
	origHostname := getHostname
	origUser := getCurrentUser
	origEnv := getEnv
	t.Cleanup(func() {
		getHostname = origHostname
		getCurrentUser = origUser
		getEnv = origEnv
	})

	getHostname = func() (string, error) { return "Host Name", nil }
	getCurrentUser = func() (*user.User, error) { return &user.User{Username: "User/Name"}, nil }
	getEnv = func(string) string { return "" }

	identity := JournalIdentity()
	if identity != "Host_Name-User_Name" {
		t.Fatalf("unexpected identity: %s", identity)
	}
}

func TestResolveJournalPathDefault(t *testing.T) {
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

	path, err := ResolveJournalPath("", "")
	if err != nil {
		t.Fatalf("ResolveJournalPath error: %v", err)
	}
	want := filepath.Join(stateHome, "gimedic", "journals", "Host-User.jsonl")
	if path != want {
		t.Fatalf("unexpected journal path: %s", path)
	}
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("journal dir not created: %v", err)
	}
}

func TestResolveJournalPathsFiltersSelf(t *testing.T) {
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

	dir := t.TempDir()
	self := filepath.Join(dir, "Host-User.jsonl")
	other := filepath.Join(dir, "Other.jsonl")
	if err := os.WriteFile(self, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write self: %v", err)
	}
	if err := os.WriteFile(other, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write other: %v", err)
	}

	paths, err := ResolveJournalPaths(dir, nil)
	if err != nil {
		t.Fatalf("ResolveJournalPaths error: %v", err)
	}
	if len(paths) != 1 || paths[0] != other {
		t.Fatalf("unexpected paths: %#v", paths)
	}
}

func TestResolveJournalPathsArgsFiltersSelf(t *testing.T) {
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

	self := "/tmp/Host-User.jsonl"
	other := "/tmp/Other.jsonl"
	paths, err := ResolveJournalPaths("", []string{self, other})
	if err != nil {
		t.Fatalf("ResolveJournalPaths error: %v", err)
	}
	if len(paths) != 1 || paths[0] != other {
		t.Fatalf("unexpected paths: %#v", paths)
	}
}
