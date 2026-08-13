package configtemplate

import (
	"fmt"
	"os"

	"github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var templateFile string

var setCmd = &cobra.Command{
	Use:     "set <server|client>",
	Short:   "Заменить шаблон содержимым файла",
	Example: "  sudo gwg template set client --file client.conf.tmpl",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		content, err := os.ReadFile(templateFile)
		if err != nil {
			return fmt.Errorf("read template file %s: %w", templateFile, err)
		}
		return server.WriteConfigTemplate(args[0], content)
	},
}

func init() {
	TemplateCmd.AddCommand(setCmd)
	setCmd.Flags().StringVarP(&templateFile, "file", "f", "", "путь к файлу шаблона")
	_ = setCmd.MarkFlagRequired("file")
}
