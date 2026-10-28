package command

import (
	"fmt"

	wireguard "github.com/PavelMilanov/go-wg-manager/modules/wireguard"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	showCmd := &cobra.Command{
		Use:     "show <server|client|gateway>",
		Short:   "Показать текущий шаблон",
		Example: "  gwu wireguard template show server",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			content, err := wireguard.ReadConfigTemplate(args[0])
			if err != nil {
				return err
			}
			fmt.Print(string(content))
			return nil
		},
	}

	return showCmd
}
