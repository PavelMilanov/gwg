package user

import (
	"github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <name>",
	Short:   "Добавить пользователя",
	Example: "  gwg user add alice",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return server.AddUser(args[0])
	},
}

func init() {
	UserCmd.AddCommand(addCmd)
}
