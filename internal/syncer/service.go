package syncer

import "time"

type Service struct {
	DBPath          string
	JournalDir      string
	InhibitDuration time.Duration
}

func (s Service) ResolveJournalPath(arg string) (string, error) {
	return ResolveJournalPath(s.JournalDir, arg)
}

func (s Service) ResolveJournalPaths(args []string) ([]string, error) {
	return ResolveJournalPaths(s.JournalDir, args)
}

func (s Service) OwnJournalPath() (string, error) {
	return OwnJournalPath(s.JournalDir)
}

func (s Service) Push(journalPath string) (int, error) {
	statePath, err := SyncStatePath(s.DBPath, journalPath)
	if err != nil {
		return 0, err
	}
	state, err := LoadSyncState(statePath)
	if err != nil {
		return 0, err
	}

	inhibited, err := ShouldInhibit(s.DBPath)
	if err != nil {
		return 0, err
	}
	if inhibited {
		return 0, nil
	}

	storage, err := LoadStorage(s.DBPath)
	if err != nil {
		return 0, err
	}

	current := SnapshotFromStorage(storage)
	localEvents := DiffSnapshots(state.Snapshot, current)
	if len(localEvents) > 0 {
		if err := AppendJournalEvents(journalPath, localEvents); err != nil {
			return 0, err
		}
	}

	state.Snapshot = current
	if err := SaveSyncState(statePath, state); err != nil {
		return 0, err
	}
	return len(localEvents), nil
}

func (s Service) Pull(journalPaths []string) (int, error) {
	appliedTotal := 0
	selfJournalPath, err := s.OwnJournalPath()
	if err != nil {
		return 0, err
	}
	for _, journalPath := range journalPaths {
		statePath, err := SyncStatePath(s.DBPath, journalPath)
		if err != nil {
			return 0, err
		}
		state, err := LoadSyncState(statePath)
		if err != nil {
			return 0, err
		}

		applied, changed, newOffset, err := ApplyJournal(s.DBPath, journalPath, state.JournalOffset)
		if err != nil {
			return 0, err
		}
		appliedTotal += applied

		if changed {
			if err := SetInhibit(s.DBPath, s.InhibitDuration); err != nil {
				return 0, err
			}
		}

		storage, err := LoadStorage(s.DBPath)
		if err != nil {
			return 0, err
		}
		state.Snapshot = SnapshotFromStorage(storage)
		state.JournalOffset = newOffset
		if err := SaveSyncState(statePath, state); err != nil {
			return 0, err
		}
	}
	if err := RefreshOwnSnapshot(s.DBPath, selfJournalPath); err != nil {
		return appliedTotal, err
	}
	return appliedTotal, nil
}
