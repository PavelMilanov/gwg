package bandwidth

import "github.com/spf13/cobra"

var BandwidthCmd = &cobra.Command{
	Use:     "bandwidth",
	Short:   "Управление классами пропускной способности",
	Long:    "Группа команд для создания, удаления и просмотра классов пропускной способности traffic control.",
	Example: "  gwg tc bandwidth list\n  gwg tc bandwidth add regular --min 2Mbit --ceil 3Mbit",
	Args:    cobra.NoArgs,
	RunE: func(command *cobra.Command, _ []string) error {
		return command.Help()
	},
}
