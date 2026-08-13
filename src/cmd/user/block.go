package user

import (
	"github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:     "block <name>",
	Short:   "Заблокировать пользователя",
	Example: "  gwg user block alice",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return server.ChangeStatusUser(args[0], "block")
	},
}

func init() {
	UserCmd.AddCommand(blockCmd)
}
