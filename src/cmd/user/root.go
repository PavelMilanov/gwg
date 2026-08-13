package user

import "github.com/spf13/cobra"

var UserCmd = &cobra.Command{
	Use:     "user",
	Short:   "Управление пользователями WireGuard",
	Long:    "Группа команд для добавления, удаления, блокировки и просмотра пользователей WireGuard.",
	Example: "  gwg user list\n  gwg user add alice\n  gwg user block alice",
	Args:    cobra.NoArgs,
	RunE: func(command *cobra.Command, _ []string) error {
		return command.Help()
	},
}
