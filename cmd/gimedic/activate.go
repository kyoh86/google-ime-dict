package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kyoh86/gimedic/internal/scheduler"
	"github.com/kyoh86/gimedic/internal/syncer"
	"github.com/spf13/cobra"
)

var activateCommand = &cobra.Command{
	Use:   "activate",
	Short: "Activate previously scheduled sync configuration",
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
		execPath, err := cmd.Flags().GetString("exec")
		if err != nil {
			return err
		}
		if execPath == "" {
			execPath, err = os.Executable()
			if err != nil || execPath == "" {
				execPath = "gimedic"
			}
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
			if err := ensureFilesExist(files); err != nil {
				return err
			}
			for path := range files {
				if err := run("launchctl", "load", path); err != nil {
					return err
				}
			}
			return nil
		case "windows":
			return activateWindows(plan)
		default:
			files := plan.SystemdFiles(home)
			if err := ensureFilesExist(files); err != nil {
				return err
			}
			if err := run("systemctl", "--user", "daemon-reload"); err != nil {
				return err
			}
			if err := run("systemctl", "--user", "enable", "--now", plan.ServiceName+"-push.timer"); err != nil {
				return err
			}
			return run("systemctl", "--user", "enable", "--now", plan.ServiceName+"-pull.timer")
		}
	},
}

func init() {
	activateCommand.Flags().Duration("interval", 5*time.Minute, "Sync interval")
	activateCommand.Flags().String("journal-dir", "", "Shared journal directory (optional)")
	activateCommand.Flags().String("path", "", "Local user_dictionary.db path (optional)")
	activateCommand.Flags().String("exec", "", "Executable path for scheduled jobs (optional)")
	facadeCommand.AddCommand(activateCommand)
}

func run(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func activateWindows(plan scheduler.Plan) error {
	dir, err := syncer.DefaultJournalDir()
	if err != nil {
		return err
	}
	scriptPath := filepath.Join(dir, "schedules", "gimedic-schedule.ps1")
	if _, err := os.Stat(scriptPath); err == nil {
		return run("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	}
	script := plan.WindowsScript()
	return run("powershell", "-ExecutionPolicy", "Bypass", "-Command", script)
}

func ensureFilesExist(files map[string]string) error {
	missing := []string{}
	for path := range files {
		if _, err := os.Stat(path); err != nil {
			missing = append(missing, path)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("missing schedule files: %s (run gimedic schedule first)", strings.Join(missing, ", "))
}
