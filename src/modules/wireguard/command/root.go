// Package command owns the WireGuard module's Cobra command tree.
package command

import (
	initcmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/init"
	servercmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/server"
	templatecmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/template"
	usercmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command/user"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	root := &cobra.Command{Use: "wireguard", Short: "WireGuard-сервер, клиенты и шаблоны", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }}
	root.AddCommand(servercmd.NewCommand(), usercmd.NewCommand(), templatecmd.NewCommand(), initcmd.NewCommand())
	return root
}
