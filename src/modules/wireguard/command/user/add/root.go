package command

import (
	wireguard "github.com/PavelMilanov/go-wg-manager/modules/wireguard"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	addCmd := &cobra.Command{
		Use:     "add <name>",
		Short:   "Добавить пользователя",
		Example: "  gwu wireguard user add alice",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return wireguard.AddUser(args[0])
		},
	}

	return addCmd
}
