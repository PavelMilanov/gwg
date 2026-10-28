package command

import (
	resetcmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/template/reset"
	setcmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/template/set"
	showcmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/template/show"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	root := &cobra.Command{
		Use:     "template",
		Short:   "Управление шаблонами WireGuard",
		Long:    "Просмотр, замена и восстановление шаблонов серверной и клиентской конфигурации.",
		Example: "  gwu wireguard template show server\n  sudo gwu wireguard template set client --file client.conf.tmpl\n  sudo gwu wireguard template reset server",
		Args:    cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}
	root.AddCommand(resetcmd.NewCommand(), setcmd.NewCommand(), showcmd.NewCommand())
	return root
}
