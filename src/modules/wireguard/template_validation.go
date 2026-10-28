package wireguard

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

// natTableName is injective for validated interface names and produces only
// nft identifier-safe characters, including for aliases containing +, = or -.
func natTableName(alias string) (string, error) {
	if err := validateInterfaceName(alias); err != nil {
		return "", err
	}
	return "gwu_nat_" + hex.EncodeToString([]byte(alias)), nil
}

func parseConfigTemplate(name string, source []byte) (*template.Template, error) {
	return template.New(name).Funcs(template.FuncMap{"natTable": natTableName}).
		Option("missingkey=error").Parse(string(source))
}

func renderConfigTemplate(tmpl *template.Template, value any) ([]byte, error) {
	var output bytes.Buffer
	if err := tmpl.Execute(&output, value); err != nil {
		// Execution errors can contain interpolated values, including private keys.
		return nil, fmt.Errorf("cannot render template %s: check field names and expressions", tmpl.Name())
	}
	if err := validateHookSyntax(output.String()); err != nil {
		return nil, fmt.Errorf("template %s: %w", tmpl.Name(), err)
	}
	return output.Bytes(), nil
}

func validateHookSyntax(rendered string) error {
	for lineNumber, line := range strings.Split(rendered, "\n") {
		key, hook, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "PreUp", "PostUp", "PreDown", "PostDown":
		default:
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, "/bin/bash", "--noprofile", "--norc", "-n")
		for _, entry := range os.Environ() {
			if !strings.HasPrefix(entry, "BASH_ENV=") && !strings.HasPrefix(entry, "ENV=") {
				cmd.Env = append(cmd.Env, entry)
			}
		}
		cmd.Stdin = strings.NewReader(hook + "\n")
		err := cmd.Run()
		cancel()
		if err != nil {
			return fmt.Errorf("cannot validate %s shell syntax at line %d (requires /bin/bash)", strings.TrimSpace(key), lineNumber+1)
		}
	}
	return nil
}

// Representative, non-secret values catch unknown fields in both Peer branches.
// This is not exhaustive evaluation of arbitrary user-defined template branches.
func (config configTemplate) validationValues() []any {
	const key = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	user := UserConfig{
		ClientPrivateKey: key, ClientPublicKey: key, ClientLocalAddress: "10.7.0.2/32",
		ServerPublicKey: key, ServerIp: "192.0.2.10", ServerPort: 51830,
		DnsResolv: "10.7.0.1", Name: "validation-client", Status: "active",
	}
	blocked := user
	blocked.Status = ""
	if config.filename == clientTemplate.filename {
		return []any{user, blocked}
	}
	server := WgServerConfig{
		ServerPrivateKey: key, ServerPublicKey: key, LocalAddress: "10.7.0.1/24",
		PublicAddress: "192.0.2.10", ListenPort: 51830, Eth: "eth0", Alias: "wg7",
		DnsResolv: "10.7.0.1", Users: []UserConfig{user, blocked},
	}
	withoutUsers := server
	withoutUsers.Users = nil
	return []any{server, withoutUsers}
}
