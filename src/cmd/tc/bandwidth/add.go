package bandwidth

import (
	traffic "github.com/PavelMilanov/go-wg-manager/tc"
	"github.com/spf13/cobra"
)

var minimum, ceiling string

var addCmd = &cobra.Command{
	Use:   "add <description>",
	Short: "Добавить класс пропускной способности",
	Example: `  gwg tc bandwidth add regular
  gwg tc bandwidth add premium --min 10Mbit --ceil 20Mbit`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return traffic.AddBandwidth(args[0], minimum, ceiling)
	},
}

func init() {
	BandwidthCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&minimum, "min", "m", "50Mbit", "минимальная скорость")
	addCmd.Flags().StringVarP(&ceiling, "ceil", "c", "950Mbit", "максимальная скорость")
}
