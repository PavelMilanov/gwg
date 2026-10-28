package command

import (
	traffic "github.com/PavelMilanov/go-wg-manager/modules/qos"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var downInterface string
	down := &cobra.Command{Use: "down", Short: "Удалить только канал с подтверждённым владением gwu", Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error { return traffic.Remove(downInterface) }}
	down.Flags().StringVar(&downInterface, "interface", "wg0", "интерфейс")
	return down
}
