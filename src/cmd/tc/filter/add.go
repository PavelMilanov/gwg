package filter

import (
	traffic "github.com/PavelMilanov/go-wg-manager/tc"
	"github.com/spf13/cobra"
)

var userName, classID string

var addCmd = &cobra.Command{
	Use:     "add <description>",
	Short:   "Добавить фильтр пользователя",
	Example: "  gwg tc filter add alice-limit --user alice --class 2",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return traffic.AddFilter(args[0], userName, classID)
	},
}

func init() {
	FilterCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&userName, "user", "u", "", "имя пользователя")
	addCmd.Flags().StringVarP(&classID, "class", "c", "", "ID класса")
	_ = addCmd.MarkFlagRequired("user")
	_ = addCmd.MarkFlagRequired("class")
}
