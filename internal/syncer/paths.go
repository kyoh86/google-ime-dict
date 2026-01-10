package syncer

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

var journalSafePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

var (
	getHostname    = os.Hostname
	getCurrentUser = user.Current
	getEnv         = os.Getenv
)

// DefaultJournalDir returns the default journal directory.
func DefaultJournalDir() (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "journals"), nil
}

// JournalIdentity returns a stable identity for naming a journal file.
func JournalIdentity() string {
	host, _ := getHostname()
	name := ""
	if current, err := getCurrentUser(); err == nil {
		name = current.Username
	}
	if name == "" {
		name = getEnv("USERNAME")
	}
	if name == "" {
		name = getEnv("USER")
	}
	host = journalSafePattern.ReplaceAllString(host, "_")
	name = journalSafePattern.ReplaceAllString(name, "_")
	host = strings.Trim(host, "_")
	name = strings.Trim(name, "_")
	if host == "" && name == "" {
		return "unknown"
	}
	if name == "" {
		return host
	}
	if host == "" {
		return name
	}
	return host + "-" + name
}

// OwnJournalFileName returns the expected journal file name for this host/user.
func OwnJournalFileName() string {
	return JournalIdentity() + ".jsonl"
}

// OwnJournalPath returns the full path for the local journal file.
func OwnJournalPath(journalDir string) (string, error) {
	dir, err := resolveJournalDir(journalDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, OwnJournalFileName()), nil
}

// ResolveJournalPath resolves the journal path with optional override.
func ResolveJournalPath(journalDir, arg string) (string, error) {
	if arg != "" {
		return arg, nil
	}
	dir, err := resolveJournalDir(journalDir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, OwnJournalFileName()), nil
}

// ResolveJournalPaths resolves journal paths, filtering the local journal.
func ResolveJournalPaths(journalDir string, args []string) ([]string, error) {
	if len(args) > 0 {
		return filterSelfJournals(args), nil
	}
	dir, err := resolveJournalDir(journalDir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	paths := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	return filterSelfJournals(paths), nil
}

func filterSelfJournals(paths []string) []string {
	own := OwnJournalFileName()
	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		if filepath.Base(path) == own {
			continue
		}
		filtered = append(filtered, path)
	}
	return filtered
}

func resolveJournalDir(journalDir string) (string, error) {
	if journalDir != "" {
		return journalDir, nil
	}
	return DefaultJournalDir()
}
