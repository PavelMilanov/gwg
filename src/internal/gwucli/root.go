package gwucli

import (
	"runtime"

	versioncmd "github.com/PavelMilanov/go-wg-manager/internal/gwucli/command/version"
	qoscmd "github.com/PavelMilanov/go-wg-manager/modules/qos/command"
	wgcmd "github.com/PavelMilanov/go-wg-manager/modules/wireguard/command"
	"github.com/spf13/cobra"
)

// NewCommand returns an independent command tree with its own flag state.
func NewCommand() *cobra.Command {
	root := &cobra.Command{
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
		Use:               "gwu",
		Short:             "Менеджер WireGuard-сервера",
		Long:              "gwu - командная утилита для настройки и администрирования WireGuard-сервера.",
		Version:           Version,
		Args:              cobra.NoArgs,
		SilenceErrors:     true,
		SilenceUsage:      true,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}
	root.SetVersionTemplate("gwu version: {{.Version}}\ngo version: " + runtime.Version() + "\n")
	root.AddCommand(wgcmd.NewCommand(), qoscmd.NewCommand(), versioncmd.NewCommand(Version))
	return root
}
