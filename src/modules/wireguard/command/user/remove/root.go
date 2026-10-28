package command

import (
	wireguard "github.com/PavelMilanov/go-wg-manager/modules/wireguard"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove <name>",
		Short:   "Удалить пользователя",
		Example: "  gwu wireguard user remove alice",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return wireguard.RemoveUser(args[0])
		},
	}

	return removeCmd
}
