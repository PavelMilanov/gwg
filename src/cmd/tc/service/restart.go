package service

import (
	traffic "github.com/PavelMilanov/go-wg-manager/tc"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:     "restart",
	Short:   "Перечитать настройки и перезапустить службу",
	Example: "  gwg tc service restart",
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return traffic.RestartService()
	},
}

func init() {
	ServiceCmd.AddCommand(restartCmd)
}
