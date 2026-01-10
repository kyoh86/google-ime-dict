package main

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var journalSafePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func defaultJournalDir() (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "journals"), nil
}

func journalIdentity() string {
	host, _ := os.Hostname()
	name := ""
	if current, err := user.Current(); err == nil {
		name = current.Username
	}
	if name == "" {
		name = os.Getenv("USERNAME")
	}
	if name == "" {
		name = os.Getenv("USER")
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

func ownJournalFileName() string {
	return journalIdentity() + ".jsonl"
}

func ownJournalPath(cmd *cobra.Command) (string, error) {
	dir, err := journalDir(cmd)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ownJournalFileName()), nil
}

func resolveJournalPath(cmd *cobra.Command, args []string) (string, error) {
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}
	dir, err := journalDir(cmd)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, ownJournalFileName()), nil
}

func resolveJournalPaths(cmd *cobra.Command, args []string) ([]string, error) {
	if len(args) > 0 {
		return filterSelfJournals(args), nil
	}
	dir, err := journalDir(cmd)
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
	own := ownJournalFileName()
	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		if filepath.Base(path) == own {
			continue
		}
		filtered = append(filtered, path)
	}
	return filtered
}

func journalDir(cmd *cobra.Command) (string, error) {
	if cmd != nil {
		if flagDir, err := cmd.Flags().GetString("journal-dir"); err == nil && flagDir != "" {
			return flagDir, nil
		}
	}
	return defaultJournalDir()
}
