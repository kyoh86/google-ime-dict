package main

import (
	"github.com/apex/log"
	"github.com/kyoh86/gimedic/internal/syncer"
	"github.com/spf13/cobra"
)

var pushCommand = &cobra.Command{
	Use:   "push [journal.jsonl]",
	Short: "Append local changes to a shared journal",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, err := resolvePath(cmd, nil)
		if err != nil {
			return err
		}
		journalDir, err := cmd.Flags().GetString("journal-dir")
		if err != nil {
			return err
		}
		service := syncer.Service{
			DBPath:     dbPath,
			JournalDir: journalDir,
		}
		journalPath, err := service.ResolveJournalPath(firstArg(args))
		if err != nil {
			return err
		}
		wrote, err := service.Push(journalPath)
		if err != nil {
			return err
		}
		if wrote > 0 {
			log.Infof("push: wrote %d events to %s", wrote, journalPath)
		}
		return nil
	},
}

func init() {
	pushCommand.Flags().String("path", "", "Local user_dictionary.db path (overrides auto-detect)")
	pushCommand.Flags().String("journal-dir", "", "Directory for journal files (overrides default)")
	facadeCommand.AddCommand(pushCommand)
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
