package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

type journalEvent struct {
	Timestamp string `json:"ts,omitempty"`
	Op        string `json:"op"`
	Dict      string `json:"dict"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Pos       int32  `json:"pos"`
	Comment   string `json:"comment,omitempty"`
	Locale    string `json:"locale,omitempty"`
}

type entryState struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
	Locale  string `json:"locale,omitempty"`
	Pos     int32  `json:"pos"`
}

type snapshot struct {
	Dictionaries map[string]map[string]entryState `json:"dictionaries"`
}

type syncState struct {
	JournalOffset int64    `json:"journal_offset"`
	Snapshot      snapshot `json:"snapshot"`
}

type dbState struct {
	InhibitUntil string `json:"inhibit_until,omitempty"`
}

func loadSyncState(path string) (syncState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return syncState{Snapshot: snapshot{Dictionaries: map[string]map[string]entryState{}}}, nil
		}
		return syncState{}, err
	}
	var state syncState
	if err := json.Unmarshal(data, &state); err != nil {
		return syncState{}, err
	}
	return state, nil
}

func saveSyncState(path string, state syncState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func syncStatePath(dbPath, journalPath string) (string, error) {
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
