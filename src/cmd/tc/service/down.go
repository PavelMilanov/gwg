package service

import (
	traffic "github.com/PavelMilanov/go-wg-manager/tc"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:     "down",
	Short:   "Остановить и удалить службу",
	Example: "  gwg tc service down",
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return traffic.DownService()
	},
}

func init() {
	ServiceCmd.AddCommand(downCmd)
}
