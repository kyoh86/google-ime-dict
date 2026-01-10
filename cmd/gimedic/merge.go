package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/kyoh86/gimedic"
	"github.com/kyoh86/gimedic/internal/syncer"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var ingestCommand = &cobra.Command{
	Use:   "ingest <from.db> [to.db]",
	Short: "Ingest entries from one dictionary file into another",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		outPath, err := cmd.Flags().GetString("out")
		if err != nil {
			return err
		}
		fromPath := args[0]
		toPath, err := resolvePath(cmd, args[1:])
		if err != nil {
			return err
		}
		if outPath == "" {
			outPath = toPath
		}
		fromStorage, err := syncer.LoadStorage(fromPath)
		if err != nil {
			return err
		}
		toStorage, err := syncer.LoadStorage(toPath)
		if err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)
		if err := applyMerge(cmd.OutOrStdout(), reader, fromStorage, toStorage); err != nil {
			return err
		}

		if outPath == toPath {
			if err := backupFile(toPath, toPath+".bak"); err != nil {
				return err
			}
		}
		return syncer.WriteStorage(outPath, toStorage)
	},
}

func init() {
	ingestCommand.Flags().String("out", "", "Output path (default: overwrite target with .bak)")
	ingestCommand.Flags().String("path", "", "Target user_dictionary.db path (overrides auto-detect)")
	facadeCommand.AddCommand(ingestCommand)
}

func backupFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func applyMerge(out io.Writer, in *bufio.Reader, fromStorage, toStorage *gimedic.UserDictionaryStorage) error {
	fromDicts := map[string]*gimedic.UserDictionary{}
	for _, d := range fromStorage.GetDictionaries() {
		fromDicts[d.GetName()] = d
	}
	toDicts := map[string]*gimedic.UserDictionary{}
	for _, d := range toStorage.GetDictionaries() {
		toDicts[d.GetName()] = d
	}

	names := make([]string, 0, len(fromDicts))
	for name := range fromDicts {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fromDict := fromDicts[name]
		toDict := toDicts[name]
		if toDict == nil {
			msg := fmt.Sprintf("ADD DICTIONARY %q (%d entries) [y/N]: ", name, len(fromDict.GetEntries()))
			ok, err := askYesNo(out, in, msg)
			if err != nil {
				return err
			}
			if ok {
				newDict := proto.Clone(fromDict).(*gimedic.UserDictionary)
				newID := syncer.UniqueDictionaryID(toStorage)
				newDict.Id = &newID
				toStorage.Dictionaries = append(toStorage.Dictionaries, newDict)
				toDicts[name] = newDict
			}
			continue
		}
		if err := mergeEntries(out, in, name, fromDict, toDict); err != nil {
			return err
		}
	}
	return nil
}

func mergeEntries(out io.Writer, in *bufio.Reader, name string, fromDict, toDict *gimedic.UserDictionary) error {
	fromEntries := map[string]*gimedic.UserDictionary_Entry{}
	for _, e := range fromDict.GetEntries() {
		fromEntries[entryKey(e)] = e
	}
	toEntries := map[string]*gimedic.UserDictionary_Entry{}
	for _, e := range toDict.GetEntries() {
		toEntries[entryKey(e)] = e
	}

	keys := make([]string, 0, len(fromEntries))
	for key := range fromEntries {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fromEntry := fromEntries[key]
		toEntry := toEntries[key]
		if toEntry == nil {
			msg := fmt.Sprintf("ADD [%s] %s [y/N]: ", name, formatEntry(fromEntry))
			ok, err := askYesNo(out, in, msg)
			if err != nil {
				return err
			}
			if ok {
				toDict.Entries = append(toDict.Entries, proto.Clone(fromEntry).(*gimedic.UserDictionary_Entry))
			}
			continue
		}
		if !entryEqual(fromEntry, toEntry) {
			msg := fmt.Sprintf("UPDATE [%s] %s -> %s [y/N]: ", name, formatEntry(toEntry), formatEntry(fromEntry))
			ok, err := askYesNo(out, in, msg)
			if err != nil {
				return err
			}
			if ok {
				toEntry.Comment = fromEntry.Comment
				toEntry.Pos = fromEntry.Pos
				toEntry.Locale = fromEntry.Locale
			}
		}
	}

	var deleteKeys []string
	for key := range toEntries {
		if _, ok := fromEntries[key]; !ok {
			deleteKeys = append(deleteKeys, key)
		}
	}
	sort.Strings(deleteKeys)
	if len(deleteKeys) == 0 {
		return nil
	}

	var kept []*gimedic.UserDictionary_Entry
	for _, entry := range toDict.GetEntries() {
		key := entryKey(entry)
		if _, ok := fromEntries[key]; ok {
			kept = append(kept, entry)
			continue
		}
		msg := fmt.Sprintf("DELETE [%s] %s [y/N]: ", name, formatEntry(entry))
		ok, err := askYesNo(out, in, msg)
		if err != nil {
			return err
		}
		if !ok {
			kept = append(kept, entry)
		}
	}
	toDict.Entries = kept
	return nil
}

func askYesNo(out io.Writer, in *bufio.Reader, msg string) (bool, error) {
	for {
		if _, err := fmt.Fprint(out, msg); err != nil {
			return false, err
		}
		line, err := in.ReadString('\n')
		if err != nil && err != io.EOF {
			return false, err
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.EqualFold(line, "n") {
			return false, nil
		}
		if strings.EqualFold(line, "y") {
			return true, nil
		}
		if err == io.EOF {
			return false, nil
		}
	}
}

func entryKey(entry *gimedic.UserDictionary_Entry) string {
	return entry.GetKey() + "\u0000" + entry.GetValue()
}

func entryEqual(a, b *gimedic.UserDictionary_Entry) bool {
	if a.GetComment() != b.GetComment() {
		return false
	}
	if a.GetPos() != b.GetPos() {
		return false
	}
	if a.GetLocale() != b.GetLocale() {
		return false
	}
	return true
}

func formatEntry(entry *gimedic.UserDictionary_Entry) string {
	pos := gimedic.Part(entry.GetPos()).String()
	comment := strings.ReplaceAll(entry.GetComment(), "\n", " ")
	locale := entry.GetLocale()
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s\t%s\t%s", entry.GetKey(), entry.GetValue(), pos)
	if comment != "" || locale != "" {
		fmt.Fprintf(&buf, "\t%s\t%s", comment, locale)
	}
	return buf.String()
}

func resolvePath(cmd *cobra.Command, args []string) (string, error) {
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}
	if flagPath, err := cmd.Flags().GetString("path"); err == nil && flagPath != "" {
		return flagPath, nil
	}
	path, candidates, err := gimedic.FindUserDictionaryPath()
	if err == nil {
		return path, nil
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("%w; provide --path", err)
	}
	return "", fmt.Errorf("%w; tried: %s (use --path to override)", err, strings.Join(candidates, ", "))
}
