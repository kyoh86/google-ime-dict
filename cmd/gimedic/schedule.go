package main

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/apex/log"
	"github.com/kyoh86/gimedic/internal/scheduler"
	"github.com/kyoh86/gimedic/internal/syncer"
	"github.com/spf13/cobra"
)

var scheduleCommand = &cobra.Command{
	Use:   "schedule",
	Short: "Generate periodic sync configuration for the current OS",
	RunE: func(cmd *cobra.Command, _ []string) error {
		interval, err := cmd.Flags().GetDuration("interval")
		if err != nil {
			return err
		}
		journalDir, err := cmd.Flags().GetString("journal-dir")
		if err != nil {
			return err
		}
		dbPath, err := cmd.Flags().GetString("path")
		if err != nil {
			return err
		}
		execPath, err := os.Executable()
		if err != nil || execPath == "" {
			execPath = "gimedic"
		}

		plan := scheduler.DefaultPlan(execPath, interval)
		plan.JournalDir = journalDir
		plan.DBPath = dbPath

		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		switch runtime.GOOS {
		case "darwin":
			files := plan.LaunchdFiles(home)
			if err := writeScheduleFiles(files); err != nil {
				return err
			}
			printScheduleSummary(files)
			return nil
		case "windows":
			path, err := writeWindowsSchedule(plan)
			if err != nil {
				return err
			}
			log.Infof("created %s", path)
			return nil
		default:
			files := plan.SystemdFiles(home)
			if err := writeScheduleFiles(files); err != nil {
				return err
			}
			printScheduleSummary(files)
			return nil
		}
	},
}

func init() {
	scheduleCommand.Flags().Duration("interval", 5*time.Minute, "Sync interval")
	scheduleCommand.Flags().String("journal-dir", "", "Shared journal directory (optional)")
	scheduleCommand.Flags().String("path", "", "Local user_dictionary.db path (optional)")
	facadeCommand.AddCommand(scheduleCommand)
}

func writeScheduleFiles(files map[string]string) error {
	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
		log.Infof("wrote %s", path)
	}
	return nil
}

func writeWindowsSchedule(plan scheduler.Plan) (string, error) {
	dir, err := syncer.DefaultJournalDir()
	if err != nil {
		return "", err
	}
	outDir := filepath.Join(dir, "schedules")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(outDir, "gimedic-schedule.ps1")
	content := plan.WindowsScript()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func printScheduleSummary(files map[string]string) {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		log.Infof("created %s", path)
	}
}
