package command

import (
	wireguard "github.com/PavelMilanov/go-wg-manager/modules/wireguard"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	blockCmd := &cobra.Command{
		Use:     "block <name>",
		Short:   "Заблокировать пользователя",
		Example: "  gwu wireguard user block alice",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return wireguard.ChangeStatusUser(args[0], "block")
		},
	}

	return blockCmd
}
