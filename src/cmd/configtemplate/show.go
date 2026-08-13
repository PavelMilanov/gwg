package configtemplate

import (
	"fmt"

	"github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show <server|client>",
	Short:   "Показать текущий шаблон",
	Example: "  gwg template show server",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		content, err := server.ReadConfigTemplate(args[0])
		if err != nil {
			return err
		}
		fmt.Print(string(content))
		return nil
	},
}

func init() {
	TemplateCmd.AddCommand(showCmd)
}
