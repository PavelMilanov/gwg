package command

import (
	"encoding/json"

	traffic "github.com/PavelMilanov/go-wg-manager/modules/qos"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var showInterface string
	show := &cobra.Command{Use: "show", Short: "Показать фактическое состояние общего канала", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			state, err := traffic.Inspect(showInterface)
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(state)
		}}
	show.Flags().StringVar(&showInterface, "interface", "wg0", "интерфейс")
	return show
}
