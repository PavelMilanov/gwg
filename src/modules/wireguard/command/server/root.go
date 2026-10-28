package command

import (
	installcmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/server/install"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	root := &cobra.Command{
		Use:     "server",
		Short:   "Управление WireGuard-сервером",
		Long:    "Установка WireGuard-сервера.",
		Example: "  gwu wireguard server install",
		Args:    cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}
	root.AddCommand(installcmd.NewCommand())
	return root
}
