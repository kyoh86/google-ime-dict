package syncer

import "time"

func ShouldInhibit(dbPath string) (bool, error) {
	statePath, err := dbStatePath(dbPath)
	if err != nil {
		return false, err
	}
	state, err := loadDBState(statePath)
	if err != nil {
		return false, err
	}
	if state.InhibitUntil == "" {
		return false, nil
	}
	until, err := time.Parse(time.RFC3339Nano, state.InhibitUntil)
	if err != nil {
		return false, nil
	}
	return time.Now().Before(until), nil
}

func SetInhibit(dbPath string, duration time.Duration) error {
	statePath, err := dbStatePath(dbPath)
	if err != nil {
		return err
	}
	state, err := loadDBState(statePath)
	if err != nil {
		return err
	}
	state.InhibitUntil = time.Now().Add(duration).Format(time.RFC3339Nano)
	return saveDBState(statePath, state)
}
