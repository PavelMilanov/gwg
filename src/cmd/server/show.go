package server

import (
	serverapi "github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "Показать состояние сервера и соединений",
	Example: "  gwg server show",
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return serverapi.ShowPeers()
	},
}

func init() {
	ServerCmd.AddCommand(showCmd)
}
