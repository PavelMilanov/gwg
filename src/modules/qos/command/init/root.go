package command

import (
	traffic "github.com/PavelMilanov/go-wg-manager/modules/qos"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var initial traffic.Config
	initialize := &cobra.Command{Use: "init", Short: "Сгенерировать конфиг и unit из .tmpl без применения QoS", Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error { return traffic.Initialize(initial) }}
	initialize.Flags().StringVar(&initial.Interface, "interface", "wg0", "интерфейс")
	initialize.Flags().StringVar(&initial.Rate, "rate", "", "общий лимит, например 75Mbit")
	_ = initialize.MarkFlagRequired("rate")
	return initialize
}
