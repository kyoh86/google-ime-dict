package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kyoh86/gimedic"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var decodeCommand = &cobra.Command{
	Use:   "decode",
	Short: "Decode a dictionary to human-readable",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolvePath(cmd, args)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		raw, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		var storage gimedic.UserDictionaryStorage
		if err := proto.Unmarshal(raw, &storage); err != nil {
			return err
		}

		fmt.Printf("version: %d\n", storage.GetVersion())
		for _, dict := range storage.GetDictionaries() {
			fmt.Printf("\n[%s] id=%d\n", dict.GetName(), dict.GetId())
			for _, entry := range dict.GetEntries() {
				pos := gimedic.Part(entry.GetPos()).String()
				comment := strings.ReplaceAll(entry.GetComment(), "\n", " ")
				fmt.Printf("%s\t%s\t%s\t%s\t%s\n",
					entry.GetKey(),
					entry.GetValue(),
					pos,
					comment,
					entry.GetLocale(),
				)
			}
		}

		return nil
	},
}

func init() {
	decodeCommand.Flags().String("path", "", "Path to user_dictionary.db (overrides auto-detect)")
	facadeCommand.AddCommand(decodeCommand)
}
