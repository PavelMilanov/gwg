package filter

import (
	traffic "github.com/PavelMilanov/go-wg-manager/tc"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <description>",
	Short:   "Удалить фильтр",
	Example: "  gwg tc filter remove alice-limit",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return traffic.RemoveFilter(args[0])
	},
}

func init() {
	FilterCmd.AddCommand(removeCmd)
}
