package command

import (
	traffic "github.com/PavelMilanov/go-wg-manager/modules/qos"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	render := &cobra.Command{Use: "render", Short: "Перегенерировать unit из локального .tmpl, без установки", Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error { return traffic.Generate() }}
	return render
}
