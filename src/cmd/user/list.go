package user

import (
	"github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "Показать настроенных пользователей",
	Example: "  gwg user list",
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return server.ListUsers()
	},
}

func init() {
	UserCmd.AddCommand(listCmd)
}
