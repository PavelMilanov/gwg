package user

import (
	"github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Удалить пользователя",
	Example: "  gwg user remove alice",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return server.RemoveUser(args[0])
	},
}

func init() {
	UserCmd.AddCommand(removeCmd)
}
