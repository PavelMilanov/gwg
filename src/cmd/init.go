package cmd

import (
	"github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Подготовить систему и создать сервер wg0",
	Long:    "Команда подготавливает системные каталоги, включает IPv4 forwarding и создает WireGuard-сервер wg0.",
	Example: "  sudo gwg init",
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return server.ConfigureSystem()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
