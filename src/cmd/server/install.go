package server

import (
	serverapi "github.com/PavelMilanov/go-wg-manager/server"
	"github.com/spf13/cobra"
)

var (
	interfaceName string
	network       string
	port          int
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Создать и запустить WireGuard-сервер",
	Long:  "Команда создает конфигурацию WireGuard-сервера и запускает соответствующую службу wg-quick.",
	Example: `  sudo gwg server install
  sudo gwg server install --name vpn7 --network 10.7.0.1/24 --port 51900`,
	Args: cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return serverapi.InstallServer(interfaceName, network, port)
	},
}

func init() {
	ServerCmd.AddCommand(installCmd)
	installCmd.Flags().StringVarP(&interfaceName, "name", "n", "wg0", "имя WireGuard-интерфейса")
	installCmd.Flags().StringVar(&network, "network", "10.0.0.1/24", "приватный IPv4-адрес и префикс")
	installCmd.Flags().IntVarP(&port, "port", "p", 51830, "UDP-порт WireGuard")
}
