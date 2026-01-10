package scheduler

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Plan struct {
	ExecPath    string
	JournalDir  string
	DBPath      string
	Interval    time.Duration
	PushLabel   string
	PullLabel   string
	ServiceName string
}

func DefaultPlan(execPath string, interval time.Duration) Plan {
	return Plan{
		ExecPath:    execPath,
		Interval:    interval,
		PushLabel:   "com.kyoh86.gimedic-push",
		PullLabel:   "com.kyoh86.gimedic-pull",
		ServiceName: "gimedic",
	}
}

func (p Plan) CommandArgs(command string) []string {
	args := []string{p.ExecPath, command}
	if p.JournalDir != "" {
		args = append(args, "--journal-dir", p.JournalDir)
	}
	if p.DBPath != "" {
		args = append(args, "--path", p.DBPath)
	}
	return args
}

func (p Plan) LaunchdFiles(home string) map[string]string {
	interval := int(p.Interval.Seconds())
	if interval <= 0 {
		interval = 300
	}
	pushPath := filepath.Join(home, "Library", "LaunchAgents", p.PushLabel+".plist")
	pullPath := filepath.Join(home, "Library", "LaunchAgents", p.PullLabel+".plist")
	return map[string]string{
		pushPath: launchdPlist(p.PushLabel, p.CommandArgs("push"), interval),
		pullPath: launchdPlist(p.PullLabel, p.CommandArgs("pull"), interval),
	}
}

func (p Plan) SystemdFiles(home string) map[string]string {
	dir := filepath.Join(home, ".config", "systemd", "user")
	pushService := filepath.Join(dir, p.ServiceName+"-push.service")
	pushTimer := filepath.Join(dir, p.ServiceName+"-push.timer")
	pullService := filepath.Join(dir, p.ServiceName+"-pull.service")
	pullTimer := filepath.Join(dir, p.ServiceName+"-pull.timer")
	interval := systemdDuration(p.Interval)
	return map[string]string{
		pushService: systemdService("Gimedic push", p.CommandArgs("push")),
		pushTimer:   systemdTimer("Gimedic push timer", interval, interval),
		pullService: systemdService("Gimedic pull", p.CommandArgs("pull")),
		pullTimer:   systemdTimer("Gimedic pull timer", interval, interval),
	}
}

func (p Plan) WindowsScript() string {
	pushCmd := windowsArgs(p.CommandArgs("push"))
	pullCmd := windowsArgs(p.CommandArgs("pull"))
	return strings.Join([]string{
		`$exe = "` + p.ExecPath + `"`,
		`$push = ` + pushCmd,
		`$pull = ` + pullCmd,
		`schtasks /Create /TN "Gimedic Push" /SC MINUTE /MO ` + intervalMinutes(p.Interval) + ` /TR $push /F`,
		`schtasks /Create /TN "Gimedic Pull" /SC MINUTE /MO ` + intervalMinutes(p.Interval) + ` /TR $pull /F`,
	}, "\r\n") + "\r\n"
}

func (p Plan) DetectOS() string {
	return runtime.GOOS
}

func launchdPlist(label string, args []string, interval int) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key><string>%s</string>
    <key>ProgramArguments</key>
    <array>
%s
    </array>
    <key>StartInterval</key><integer>%d</integer>
    <key>RunAtLoad</key><true/>
  </dict>
</plist>
`, label, launchdArgs(args), interval)
}

func launchdArgs(args []string) string {
	lines := make([]string, 0, len(args))
	for _, arg := range args {
		lines = append(lines, fmt.Sprintf("      <string>%s</string>", arg))
	}
	return strings.Join(lines, "\n")
}

func systemdService(description string, args []string) string {
	return fmt.Sprintf(`[Unit]
Description=%s

[Service]
Type=oneshot
ExecStart=%s
`, description, strings.Join(args, " "))
}

func systemdTimer(description, onBoot, onActive string) string {
	return fmt.Sprintf(`[Unit]
Description=%s

[Timer]
OnBootSec=%s
OnUnitActiveSec=%s

[Install]
WantedBy=timers.target
`, description, onBoot, onActive)
}

func systemdDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds <= 0 {
		seconds = 300
	}
	return fmt.Sprintf("%ds", seconds)
}

func intervalMinutes(d time.Duration) string {
	minutes := int(d.Minutes())
	if minutes <= 0 {
		minutes = 5
	}
	return fmt.Sprintf("%d", minutes)
}

func windowsArgs(args []string) string {
	escaped := make([]string, 0, len(args))
	for _, arg := range args {
		escaped = append(escaped, `"`+arg+`"`)
	}
	return strings.Join(escaped, " ")
}
