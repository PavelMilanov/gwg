package filter

import "github.com/spf13/cobra"

var FilterCmd = &cobra.Command{
	Use:     "filter",
	Short:   "Управление фильтрами пользователей",
	Long:    "Группа команд для создания, удаления и просмотра правил классификации пользовательского трафика.",
	Example: "  gwg tc filter list\n  gwg tc filter add alice-limit --user alice --class 2",
	Args:    cobra.NoArgs,
	RunE: func(command *cobra.Command, _ []string) error {
		return command.Help()
	},
}
