package command

import (
	addcmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/user/add"
	blockcmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/user/block"
	listcmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/user/list"
	removecmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/user/remove"
	unblockcmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/user/unblock"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	root := &cobra.Command{
		Use:     "user",
		Short:   "Управление пользователями WireGuard",
		Long:    "Группа команд для добавления, удаления, блокировки и просмотра пользователей WireGuard.",
		Example: "  gwu wireguard user list\n  gwu wireguard user add alice\n  gwu wireguard user block alice",
		Args:    cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}
	root.AddCommand(addcmd.NewCommand(), blockcmd.NewCommand(), listcmd.NewCommand(), removecmd.NewCommand(), unblockcmd.NewCommand())
	return root
}
