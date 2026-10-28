package command

import (
	"fmt"

	traffic "github.com/PavelMilanov/go-wg-manager/modules/qos"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var configPath string
	apply := &cobra.Command{Use: "apply", Short: "Создать канал или изменить общую скорость без удаления root qdisc", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := traffic.ReadConfig(configPath)
			if err != nil {
				return err
			}
			if err := traffic.Apply(c); err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Shared channel applied")
			return err
		}}
	apply.Flags().StringVar(&configPath, "config", traffic.DefaultConfigPath, "JSON-конфигурация общего канала")
	return apply
}
