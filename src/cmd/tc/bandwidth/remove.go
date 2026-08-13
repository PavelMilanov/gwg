package bandwidth

import (
	traffic "github.com/PavelMilanov/go-wg-manager/tc"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <class-id>",
	Short:   "Удалить класс пропускной способности",
	Example: "  gwg tc bandwidth remove 2",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return traffic.RemoveBandwidth(args[0])
	},
}

func init() {
	BandwidthCmd.AddCommand(removeCmd)
}
