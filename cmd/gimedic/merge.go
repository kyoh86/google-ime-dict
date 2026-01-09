package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/kyoh86/gimedic"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var mergeCommand = &cobra.Command{
	Use:   "merge <from.db> <to.db>",
	Short: "Merge entries from one dictionary file into another",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		outPath, err := cmd.Flags().GetString("out")
		if err != nil {
			return err
		}
		fromPath := args[0]
		toPath := args[1]
		if outPath == "" {
			outPath = toPath
		}
		fromStorage, err := loadStorage(fromPath)
		if err != nil {
			return err
		}
		toStorage, err := loadStorage(toPath)
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
		return writeStorage(outPath, toStorage)
	},
}

func init() {
	mergeCommand.Flags().String("out", "", "Output path (default: overwrite target with .bak)")
	facadeCommand.AddCommand(mergeCommand)
}

func loadStorage(path string) (*gimedic.UserDictionaryStorage, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var storage gimedic.UserDictionaryStorage
	if err := proto.Unmarshal(raw, &storage); err != nil {
		return nil, err
	}
	return &storage, nil
}

func writeStorage(path string, storage *gimedic.UserDictionaryStorage) error {
	raw, err := proto.Marshal(storage)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
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
				newID := uniqueDictionaryID(toStorage)
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

func uniqueDictionaryID(storage *gimedic.UserDictionaryStorage) uint64 {
	used := map[uint64]struct{}{}
	for _, d := range storage.GetDictionaries() {
		used[d.GetId()] = struct{}{}
	}
	for {
		id := randomUint64()
		if id == 0 {
			continue
		}
		if _, ok := used[id]; !ok {
			return id
		}
	}
}

func randomUint64() uint64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0
	}
	return binary.LittleEndian.Uint64(b[:])
}
