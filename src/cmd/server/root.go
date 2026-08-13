package server

import "github.com/spf13/cobra"

var ServerCmd = &cobra.Command{
	Use:     "server",
	Short:   "Управление WireGuard-сервером",
	Long:    "Группа команд для установки сервера, просмотра его состояния и статистики соединений.",
	Example: "  gwg server install\n  gwg server show\n  gwg server stat",
	Args:    cobra.NoArgs,
	RunE: func(command *cobra.Command, _ []string) error {
		return command.Help()
	},
}
