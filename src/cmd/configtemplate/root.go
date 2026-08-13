package configtemplate

import "github.com/spf13/cobra"

var TemplateCmd = &cobra.Command{
	Use:     "template",
	Short:   "Управление шаблонами WireGuard",
	Long:    "Просмотр, замена и восстановление шаблонов серверной и клиентской конфигурации.",
	Example: "  gwg template show server\n  sudo gwg template set client --file client.conf.tmpl\n  sudo gwg template reset server",
	Args:    cobra.NoArgs,
	RunE: func(command *cobra.Command, _ []string) error {
		return command.Help()
	},
}
