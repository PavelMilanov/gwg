package service

import (
	traffic "github.com/PavelMilanov/go-wg-manager/tc"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "Показать конфигурацию службы",
	Example: "  gwg tc service show",
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return traffic.ShowService()
	},
}

func init() {
	ServiceCmd.AddCommand(showCmd)
}
