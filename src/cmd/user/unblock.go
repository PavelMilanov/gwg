package user

import (
	"github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var unblockCmd = &cobra.Command{
	Use:     "unblock <name>",
	Short:   "Разблокировать пользователя",
	Example: "  gwg user unblock alice",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return server.ChangeStatusUser(args[0], "unblock")
	},
}

func init() {
	UserCmd.AddCommand(unblockCmd)
}
