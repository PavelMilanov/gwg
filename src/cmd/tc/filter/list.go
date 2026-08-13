package filter

import (
	traffic "github.com/PavelMilanov/go-wg-manager/tc"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "Показать фильтры",
	Example: "  gwg tc filter list",
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return traffic.ShowFilter()
	},
}

func init() {
	FilterCmd.AddCommand(listCmd)
}
