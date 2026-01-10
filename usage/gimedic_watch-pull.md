## gimedic watch-pull

Continuously apply shared journal entries to local dictionary

```
gimedic watch-pull [journal.jsonl...] [flags]
```

### Options

```
  -h, --help                   help for watch-pull
      --inhibit-seconds int    Seconds to inhibit push after applying changes (default 2)
      --interval-seconds int   Polling interval in seconds (default 5)
      --journal-dir string     Directory for journal files (overrides default)
      --path string            Local user_dictionary.db path (overrides auto-detect)
```

### SEE ALSO

* [gimedic](gimedic.md)	 - A tool to parse user dictionary for Google IME

