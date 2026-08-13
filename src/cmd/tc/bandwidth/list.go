package bandwidth

import (
	traffic "github.com/PavelMilanov/go-wg-manager/tc"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "Показать классы пропускной способности",
	Example: "  gwg tc bandwidth list",
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return traffic.ShowBandwidth()
	},
}

func init() {
	BandwidthCmd.AddCommand(listCmd)
}
