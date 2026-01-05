package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kyoh86/gimedic"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var decodeCommand = &cobra.Command{
	Use:   "decode",
	Short: "Decode a dictionary to human-readable",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer file.Close()
		raw, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		var dict gimedic.UserDictionary
		if err := proto.Unmarshal(raw, &dict); err != nil {
			return err
		}

		fmt.Printf("%#v\n", &dict)

		return nil
	},
}

func init() {
	facadeCommand.AddCommand(decodeCommand)
}
