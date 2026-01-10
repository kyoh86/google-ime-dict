# gimedic

CLI tools for parsing and syncing Google IME user dictionaries.

[![PkgGoDev](https://pkg.go.dev/badge/kyoh86/gimedic)](https://pkg.go.dev/kyoh86/gimedic)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyoh86/gimedic)](https://goreportcard.com/report/github.com/kyoh86/gimedic)
[![Release](https://github.com/kyoh86/gimedic/workflows/Release/badge.svg)](https://github.com/kyoh86/gimedic/releases)

## Description

```console
$ gimedic man
```

`gimedic` provides tools to read and synchronize Google IME user dictionaries.

## Install

### For Golang developers

```console
$ go get github.com/kyoh86/gimedic/cmd/gimedic
```

### Homebrew/Linuxbrew

```console
$ brew tap kyoh86/tap
$ brew update
$ brew install kyoh86/tap/gimedic
```

### Makepkg

```console
$ mkdir -p gimedic_build && \
  cd gimedic_build && \
  curl -iL --fail --silent https://github.com/kyoh86/gimedic/releases/latest/download/gimedic_PKGBUILD.tar.gz | tar -xvz
$ makepkg -i
```

## Available commands

Use `gimedic [command] --help` for more information about a command.
Or see the manual in [usage/gimedic.md](./usage/gimedic.md).

## Commands

Manual: [usage/gimedic.md](./usage/gimedic.md).

## Journal Sync (Optional)

`gimedic` can synchronize Google IME user dictionaries across machines using journal files.
By default, journals live under the OS state directory (no extra flags required). Use
`--journal-dir` only when you want a shared location (e.g., Google Drive).

Quick start (recommended):

```console
$ gimedic schedule --interval 5m --journal-dir "/path/to/shared/journals"
$ gimedic activate --interval 5m --journal-dir "/path/to/shared/journals"
```

If you installed `gimedic` into a non-standard location, pass `--exec`:

```console
$ gimedic schedule --exec "/path/to/gimedic" --interval 5m
$ gimedic activate --exec "/path/to/gimedic" --interval 5m
```

Note: `--journal-dir` is optional. If you omit it, `gimedic` uses the default per-OS state directory.

Manual usage:

```console
$ gimedic push
$ gimedic pull
```

With a shared folder:

```console
$ gimedic push --journal-dir "/path/to/shared/journals"
$ gimedic pull --journal-dir "/path/to/shared/journals"
```

## Scheduled Sync Templates (Manual)

The simplest cross-OS approach is to schedule `push`/`pull` every few minutes with the OS
scheduler. **`--journal-dir` is optional**; omit it to use the default per-OS location.

### macOS (launchd)

`~/Library/LaunchAgents/com.kyoh86.gimedic-push.plist`
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key><string>com.kyoh86.gimedic-push</string>
    <key>ProgramArguments</key>
    <array>
      <string>/usr/local/bin/gimedic</string>
      <string>push</string>
      <string>--journal-dir</string>
      <string>/path/to/shared/journals</string>
    </array>
    <key>StartInterval</key><integer>300</integer>
    <key>RunAtLoad</key><true/>
  </dict>
</plist>
```

`~/Library/LaunchAgents/com.kyoh86.gimedic-pull.plist`
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key><string>com.kyoh86.gimedic-pull</string>
    <key>ProgramArguments</key>
    <array>
      <string>/usr/local/bin/gimedic</string>
      <string>pull</string>
      <string>--journal-dir</string>
      <string>/path/to/shared/journals</string>
    </array>
    <key>StartInterval</key><integer>300</integer>
    <key>RunAtLoad</key><true/>
  </dict>
</plist>
```

Enable:
```console
$ launchctl load ~/Library/LaunchAgents/com.kyoh86.gimedic-push.plist
$ launchctl load ~/Library/LaunchAgents/com.kyoh86.gimedic-pull.plist
```

### Linux (systemd --user)

`~/.config/systemd/user/gimedic-push.service`
```ini
[Unit]
Description=Gimedic push

[Service]
Type=oneshot
ExecStart=/usr/local/bin/gimedic push --journal-dir /path/to/shared/journals
```

`~/.config/systemd/user/gimedic-push.timer`
```ini
[Unit]
Description=Gimedic push timer

[Timer]
OnBootSec=1min
OnUnitActiveSec=5min

[Install]
WantedBy=timers.target
```

`~/.config/systemd/user/gimedic-pull.service`
```ini
[Unit]
Description=Gimedic pull

[Service]
Type=oneshot
ExecStart=/usr/local/bin/gimedic pull --journal-dir /path/to/shared/journals
```

`~/.config/systemd/user/gimedic-pull.timer`
```ini
[Unit]
Description=Gimedic pull timer

[Timer]
OnBootSec=2min
OnUnitActiveSec=5min

[Install]
WantedBy=timers.target
```

Enable:
```console
$ systemctl --user daemon-reload
$ systemctl --user enable --now gimedic-push.timer
$ systemctl --user enable --now gimedic-pull.timer
```

### Windows (Task Scheduler)

```powershell
$exe = "C:\path\to\gimedic.exe"
$dir = "C:\path\to\shared\journals"

schtasks /Create /TN "Gimedic Push" /SC MINUTE /MO 5 /TR "`"$exe`" push --journal-dir `"$dir`"" /F
schtasks /Create /TN "Gimedic Pull" /SC MINUTE /MO 5 /TR "`"$exe`" pull --journal-dir `"$dir`"" /F
```

### cron (macOS/Linux)

```cron
*/5 * * * * /usr/local/bin/gimedic push --journal-dir "/path/to/shared/journals"
*/5 * * * * /usr/local/bin/gimedic pull --journal-dir "/path/to/shared/journals"
```

# LICENSE

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)](http://www.opensource.org/licenses/MIT)

This software is released under the [MIT License](http://www.opensource.org/licenses/MIT), see LICENSE.
