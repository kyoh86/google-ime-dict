package syncer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

type JournalEvent struct {
	Timestamp string `json:"ts,omitempty"`
	Op        string `json:"op"`
	Dict      string `json:"dict"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Pos       int32  `json:"pos"`
	Comment   string `json:"comment,omitempty"`
	Locale    string `json:"locale,omitempty"`
}

type EntryState struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
	Locale  string `json:"locale,omitempty"`
	Pos     int32  `json:"pos"`
}

type Snapshot struct {
	Dictionaries map[string]map[string]EntryState `json:"dictionaries"`
}

type SyncState struct {
	JournalOffset int64    `json:"journal_offset"`
	Snapshot      Snapshot `json:"snapshot"`
}

type dbState struct {
	InhibitUntil string `json:"inhibit_until,omitempty"`
}

func LoadSyncState(path string) (SyncState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SyncState{Snapshot: Snapshot{Dictionaries: map[string]map[string]EntryState{}}}, nil
		}
		return SyncState{}, err
	}
	var state SyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return SyncState{}, err
	}
	return state, nil
}

func SaveSyncState(path string, state SyncState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func SyncStatePath(dbPath, journalPath string) (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(dbPath + "|" + journalPath))
	return filepath.Join(dir, "sync_"+hex.EncodeToString(sum[:])+".json"), nil
}

func dbStatePath(dbPath string) (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(dbPath))
	return filepath.Join(dir, "db_"+hex.EncodeToString(sum[:])+".json"), nil
}

func loadDBState(path string) (dbState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return dbState{}, nil
		}
		return dbState{}, err
	}
	var state dbState
	if err := json.Unmarshal(data, &state); err != nil {
		return dbState{}, err
	}
	return state, nil
}

func saveDBState(path string, state dbState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func stateDir() (string, error) {
	if env := os.Getenv("XDG_STATE_HOME"); env != "" {
		return filepath.Join(env, "gimedic"), nil
	}
	switch runtime.GOOS {
	case "windows":
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "gimedic"), nil
		}
		if config, err := os.UserConfigDir(); err == nil {
			return filepath.Join(config, "gimedic"), nil
		}
		return "", errors.New("cannot resolve state directory")
	case "darwin":
		if config, err := os.UserConfigDir(); err == nil {
			return filepath.Join(config, "gimedic"), nil
		}
		return "", errors.New("cannot resolve state directory")
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "state", "gimedic"), nil
	}
}
