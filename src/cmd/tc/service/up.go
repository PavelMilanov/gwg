package service

import (
	traffic "github.com/PavelMilanov/go-wg-manager/tc"
	"github.com/spf13/cobra"
)

var speed, maxSpeed string

var upCmd = &cobra.Command{
	Use:     "up",
	Short:   "Создать и запустить службу",
	Example: "  gwg tc service up --speed 50Mbit --max-speed 100Mbit",
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return traffic.UpService(speed, maxSpeed)
	},
}

func init() {
	ServiceCmd.AddCommand(upCmd)
	upCmd.Flags().StringVarP(&speed, "speed", "s", "", "гарантированная скорость")
	upCmd.Flags().StringVarP(&maxSpeed, "max-speed", "m", "", "максимальная скорость")
	_ = upCmd.MarkFlagRequired("max-speed")
}
