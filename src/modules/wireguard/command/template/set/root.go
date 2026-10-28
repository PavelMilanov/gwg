package command

import (
	"fmt"
	"os"

	wireguard "github.com/PavelMilanov/go-wg-manager/modules/wireguard"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var templateFile string

	setCmd := &cobra.Command{
		Use:     "set <server|client|gateway>",
		Short:   "Заменить шаблон содержимым файла",
		Example: "  sudo gwu wireguard template set client --file client.conf.tmpl",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			content, err := os.ReadFile(templateFile)
			if err != nil {
				return fmt.Errorf("read template file %s: %w", templateFile, err)
			}
			return wireguard.WriteConfigTemplate(args[0], content)
		},
	}
	setCmd.Flags().StringVarP(&templateFile, "file", "f", "", "путь к файлу шаблона")
	_ = setCmd.MarkFlagRequired("file")
	return setCmd
}
