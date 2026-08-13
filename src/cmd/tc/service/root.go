package service

import "github.com/spf13/cobra"

var ServiceCmd = &cobra.Command{
	Use:     "service",
	Short:   "Управление службой traffic control",
	Long:    "Группа команд для запуска, остановки, перезапуска и просмотра службы traffic control.",
	Example: "  gwg tc service up --max-speed 100Mbit\n  gwg tc service show",
	Args:    cobra.NoArgs,
	RunE: func(command *cobra.Command, _ []string) error {
		return command.Help()
	},
}
