package wireguard

import (
	"fmt"
	"strings"

	"github.com/PavelMilanov/go-wg-manager/internal/command"
)

var commandRunner command.Runner = command.ExecRunner{}

func generateKeys() (string, string, error) {
	privateOutput, err := commandRunner.Output("wg", "genkey")
	if err != nil {
		return "", "", fmt.Errorf("generate WireGuard private key: %w", err)
	}
	privateKey := strings.TrimSpace(string(privateOutput))
	if err := validateKey("generated private key", privateKey); err != nil {
		return "", "", err
	}

	publicOutput, err := commandRunner.OutputWithInput([]byte(privateKey+"\n"), "wg", "pubkey")
	if err != nil {
		return "", "", fmt.Errorf("generate WireGuard public key: %w", err)
	}
	publicKey := strings.TrimSpace(string(publicOutput))
	if err := validateKey("generated public key", publicKey); err != nil {
		return "", "", err
	}
	return privateKey, publicKey, nil
}

func commandServer(action string, alias string) error {
	if err := validateInterfaceName(alias); err != nil {
		return err
	}
	switch action {
	case "enable", "start", "restart":
	default:
		return fmt.Errorf("unsupported server action %q", action)
	}
	service := "wg-quick@" + alias + ".service"
	if err := commandRunner.Run("sudo", "systemctl", action, service); err != nil {
		return fmt.Errorf("%s %s: %w", action, service, err)
	}
	return nil
}
