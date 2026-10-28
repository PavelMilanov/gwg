package command

import (
	wireguard "github.com/PavelMilanov/go-wg-manager/modules/wireguard"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	unblockCmd := &cobra.Command{
		Use:     "unblock <name>",
		Short:   "Разблокировать пользователя",
		Example: "  gwu wireguard user unblock alice",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return wireguard.ChangeStatusUser(args[0], "unblock")
		},
	}

	return unblockCmd
}
