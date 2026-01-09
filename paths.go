package gimedic

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// FindUserDictionaryPath returns the first existing user_dictionary.db path.
func FindUserDictionaryPath() (string, []string, error) {
	candidates, err := userDictionaryCandidates()
	if err != nil {
		return "", nil, err
	}
	for _, path := range candidates {
		if _, statErr := os.Stat(path); statErr == nil {
			return path, candidates, nil
		}
	}
	return "", candidates, errors.New("user_dictionary.db not found")
}

func userDictionaryCandidates() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, "Library", "Application Support", "Google", "JapaneseInput", "user_dictionary.db"),
			filepath.Join(home, "Library", "Application Support", "Google", "Google Japanese Input", "user_dictionary.db"),
		}, nil
	case "windows":
		var base string
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			base = localLowFromLocal(localAppData)
		} else if appData := os.Getenv("APPDATA"); appData != "" {
			base = localLowFromRoaming(appData)
		}
		if base == "" {
			return nil, errors.New("cannot resolve LocalLow directory")
		}
		return []string{
			filepath.Join(base, "Google", "Google Japanese Input", "user_dictionary.db"),
			filepath.Join(base, "Google", "JapaneseInput", "user_dictionary.db"),
		}, nil
	default:
		return []string{
			filepath.Join(home, ".mozc", "user_dictionary.db"),
			filepath.Join(home, ".config", "google-japanese-input", "user_dictionary.db"),
		}, nil
	}
}

func localLowFromLocal(localAppData string) string {
	lower := strings.ToLower(localAppData)
	needle := strings.ToLower(filepath.Join("AppData", "Local"))
	index := strings.LastIndex(lower, needle)
	if index == -1 {
		return filepath.Join(localAppData, "..", "LocalLow")
	}
	return filepath.Join(localAppData[:index], "AppData", "LocalLow")
}

func localLowFromRoaming(appData string) string {
	lower := strings.ToLower(appData)
	needle := strings.ToLower(filepath.Join("AppData", "Roaming"))
	index := strings.LastIndex(lower, needle)
	if index == -1 {
		return filepath.Join(appData, "..", "LocalLow")
	}
	return filepath.Join(appData[:index], "AppData", "LocalLow")
}
