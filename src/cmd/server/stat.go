package server

import (
	serverapi "github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var statCmd = &cobra.Command{
	Use:     "stat",
	Short:   "Показать статистику пользователей",
	Example: "  gwg server stat",
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return serverapi.ReadWgDump()
	},
}

func init() {
	ServerCmd.AddCommand(statCmd)
}
