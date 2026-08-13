package cmd

import (
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Показать версии gwg и Go",
	Example: "  gwg version",
	Args:    cobra.NoArgs,
	RunE: func(command *cobra.Command, _ []string) error {
		command.Printf("gwg version: %s\ngo version: %s\n", Version, runtime.Version())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
