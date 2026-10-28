package command

import (
	wireguard "github.com/PavelMilanov/go-wg-manager/modules/wireguard"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	resetCmd := &cobra.Command{
		Use:     "reset <server|client|gateway>",
		Short:   "Восстановить встроенный шаблон",
		Example: "  sudo gwu wireguard template reset server",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return wireguard.ResetConfigTemplate(args[0])
		},
	}

	return resetCmd
}
