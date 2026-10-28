package command

import (
	applycmd "github.com/PavelMilanov/go-wg-manager/modules/qos/command/apply"
	downcmd "github.com/PavelMilanov/go-wg-manager/modules/qos/command/down"
	initcmd "github.com/PavelMilanov/go-wg-manager/modules/qos/command/init"
	rendercmd "github.com/PavelMilanov/go-wg-manager/modules/qos/command/render"
	showcmd "github.com/PavelMilanov/go-wg-manager/modules/qos/command/show"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "qos",
		Short: "Общий egress-канал через Netlink (прототип)",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	root.AddCommand(initcmd.NewCommand(), rendercmd.NewCommand(), applycmd.NewCommand(), showcmd.NewCommand(), downcmd.NewCommand())
	return root
}
