package command

import (
	wireguard "github.com/PavelMilanov/go-wg-manager/modules/wireguard"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "init",
		Short:   "Подготовить систему и создать сервер wg0",
		Long:    "Команда подготавливает системные каталоги, включает IPv4 forwarding и создает WireGuard-сервер wg0.",
		Example: "  sudo gwu wireguard init",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return wireguard.ConfigureSystem()
		},
	}
}
