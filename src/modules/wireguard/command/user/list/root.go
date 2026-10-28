package command

import (
	wireguard "github.com/PavelMilanov/go-wg-manager/modules/wireguard"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "Показать настроенных пользователей",
		Example: "  gwu wireguard user list",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return wireguard.ListUsers()
		},
	}

	return listCmd
}
