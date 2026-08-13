package tc

import (
	"github.com/PavelMilanov/go-wg-manager/cmd/tc/bandwidth"
	"github.com/PavelMilanov/go-wg-manager/cmd/tc/filter"
	"github.com/PavelMilanov/go-wg-manager/cmd/tc/service"
	"github.com/spf13/cobra"
)

var TCCmd = &cobra.Command{
	Use:     "tc",
	Short:   "Управление ограничениями трафика",
	Long:    "Группа команд для управления службой traffic control, классами пропускной способности и фильтрами пользователей.",
	Example: "  gwg tc service show\n  gwg tc bandwidth list\n  gwg tc filter list",
	Args:    cobra.NoArgs,
	RunE: func(command *cobra.Command, _ []string) error {
		return command.Help()
	},
}

func init() {
	TCCmd.AddCommand(service.ServiceCmd)
	TCCmd.AddCommand(bandwidth.BandwidthCmd)
	TCCmd.AddCommand(filter.FilterCmd)
}
