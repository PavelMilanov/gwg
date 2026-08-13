package configtemplate

import (
	"github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:     "reset <server|client>",
	Short:   "Восстановить встроенный шаблон",
	Example: "  sudo gwg template reset server",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return server.ResetConfigTemplate(args[0])
	},
}

func init() {
	TemplateCmd.AddCommand(resetCmd)
}
