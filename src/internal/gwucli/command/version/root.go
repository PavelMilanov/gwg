package command

import (
	"runtime"

	"github.com/spf13/cobra"
)

func NewCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Показать версии gwu и Go",
		Example: "  gwu version",
		Args:    cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			command.Printf("gwu version: %s\ngo version: %s\n", version, runtime.Version())
			return nil
		},
	}
}
