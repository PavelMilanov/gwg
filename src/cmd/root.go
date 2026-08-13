package cmd

import (
	"runtime"

	templatecmd "github.com/PavelMilanov/go-wg-manager/cmd/configtemplate"
	servercmd "github.com/PavelMilanov/go-wg-manager/cmd/server"
	trafficcmd "github.com/PavelMilanov/go-wg-manager/cmd/tc"
	usercmd "github.com/PavelMilanov/go-wg-manager/cmd/user"
	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:           "gwg",
	Short:         "Менеджер WireGuard-сервера",
	Long:          "gwg - командная утилита для настройки и администрирования WireGuard-сервера.",
	Version:       Version,
	Args:          cobra.NoArgs,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(command *cobra.Command, _ []string) error {
		return command.Help()
	},
}

func Execute() error {
	rootCmd.SetVersionTemplate("gwg version: {{.Version}}\ngo version: " + runtime.Version() + "\n")
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(servercmd.ServerCmd)
	rootCmd.AddCommand(templatecmd.TemplateCmd)
	rootCmd.AddCommand(usercmd.UserCmd)
	rootCmd.AddCommand(trafficcmd.TCCmd)
}
