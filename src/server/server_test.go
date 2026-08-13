package server

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const testKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

type fakeRunner struct {
	outputs map[string][]byte
	calls   []string
}

func (f *fakeRunner) Output(name string, args ...string) ([]byte, error) {
	key := strings.Join(append([]string{name}, args...), " ")
	f.calls = append(f.calls, key)
	out, ok := f.outputs[key]
	if !ok {
		return nil, errors.New("unexpected command: " + key)
	}
	return out, nil
}

func (f *fakeRunner) OutputWithInput(input []byte, name string, args ...string) ([]byte, error) {
	key := strings.Join(append([]string{name}, args...), " ")
	f.calls = append(f.calls, key+" < "+strings.TrimSpace(string(input)))
	out, ok := f.outputs[key]
	if !ok {
		return nil, errors.New("unexpected command: " + key)
	}
	return out, nil
}

func (f *fakeRunner) Run(name string, args ...string) error {
	f.calls = append(f.calls, strings.Join(append([]string{name}, args...), " "))
	return nil
}

func TestValidationRejectsUnsafeNames(t *testing.T) {
	for _, name := range []string{"", "../admin", "wg0;id", "name with spaces", "/tmp/x"} {
		if err := validateUserName(name); err == nil {
			t.Errorf("validateUserName(%q) accepted unsafe value", name)
		}
	}
	for _, name := range []string{"wg0;id", "../../x", "interface-name-is-too-long"} {
		if err := validateInterfaceName(name); err == nil {
			t.Errorf("validateInterfaceName(%q) accepted unsafe value", name)
		}
	}
}

func TestSetClientIPReusesGapInPrefix(t *testing.T) {
	server := WgServerConfig{LocalAddress: "10.42.1.1/24"}
	users := []UserConfig{
		{Name: "one", ClientLocalAddress: "10.42.1.2/32"},
		{Name: "three", ClientLocalAddress: "10.42.1.4/32"},
	}
	got, err := setClientIP(server, users)
	if err != nil {
		t.Fatal(err)
	}
	if got != "10.42.1.3/32" {
		t.Fatalf("address = %q, want 10.42.1.3/32", got)
	}
}

func TestSetClientIPReportsExhaustedPrefix(t *testing.T) {
	server := WgServerConfig{LocalAddress: "192.0.2.1/30"}
	users := []UserConfig{{Name: "only", ClientLocalAddress: "192.0.2.2/32"}}
	if _, err := setClientIP(server, users); err == nil {
		t.Fatal("expected exhausted-prefix error")
	}
}

func TestParseServerPrefixRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"999.1.1.1/24", "10.0.0.1/99", "10.0.0.1/31", "text10.0.0.1/24"} {
		if _, err := parseServerPrefix(value); err == nil {
			t.Errorf("parseServerPrefix(%q) succeeded", value)
		}
	}
}

func TestParseDump(t *testing.T) {
	data := []byte("priv\tpub\t51820\toff\nclient-key\tpsk\tendpoint\t10.0.0.2/32\t1\t4294967296\t7\toff\n")
	got, err := parseDump(data, []UserConfig{{Name: "alice", ClientPublicKey: "client-key"}})
	if err != nil {
		t.Fatal(err)
	}
	want := []WireguardDump{{user: "alice", ip: "10.0.0.2/32", rateRx: 4294967296, rateTx: 7}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dump = %#v, want %#v", got, want)
	}
}

func TestParseDumpRejectsMalformedLine(t *testing.T) {
	if _, err := parseDump([]byte("header\nmissing\tfields\n"), nil); err == nil {
		t.Fatal("expected malformed-dump error")
	}
}

func TestGenerateKeysUsesPipesWithoutTemporaryFiles(t *testing.T) {
	fake := &fakeRunner{outputs: map[string][]byte{
		"wg genkey": []byte(testKey + "\n"),
		"wg pubkey": []byte(testKey + "\n"),
	}}
	previous := commandRunner
	commandRunner = fake
	t.Cleanup(func() { commandRunner = previous })

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	privateKey, publicKey, err := generateKeys()
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if privateKey != testKey || publicKey != testKey {
		t.Fatalf("keys = %q, %q", privateKey, publicKey)
	}
	if after != cwd {
		t.Fatalf("working directory changed from %s to %s", cwd, after)
	}
	if got := fake.calls[1]; got != "wg pubkey < "+testKey {
		t.Fatalf("public-key call = %q", got)
	}
}

func TestConfigurationRoundTrip(t *testing.T) {
	configureTestDirectories(t)

	serverConfig := validTestServerConfig()
	if err := serverConfig.createServerConfigFile(); err != nil {
		t.Fatal(err)
	}
	userConfig := validTestUserConfig("alice", "10.7.0.2/32")
	if err := userConfig.addConfigUser("alice"); err != nil {
		t.Fatal(err)
	}
	gotServer, err := ReadServerConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	gotUsers, err := ReadClientConfigFiles()
	if err != nil {
		t.Fatal(err)
	}
	if gotServer.Alias != "wg7" || len(gotUsers) != 1 || gotUsers[0].Name != "alice" {
		t.Fatalf("unexpected round trip: %#v, %#v", gotServer, gotUsers)
	}
	info, err := os.Stat(filepath.Join(managerDir, "wg7.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("server metadata permissions = %o", info.Mode().Perm())
	}
}

func TestAddUserUpdatesConfigsAndRestartsSelectedInterface(t *testing.T) {
	configureTestDirectories(t)
	serverConfig := validTestServerConfig()
	if err := serverConfig.createServerConfigFile(); err != nil {
		t.Fatal(err)
	}
	fake := &fakeRunner{outputs: map[string][]byte{
		"wg genkey": []byte(testKey + "\n"),
		"wg pubkey": []byte(testKey + "\n"),
	}}
	previous := commandRunner
	commandRunner = fake
	t.Cleanup(func() { commandRunner = previous })

	if err := AddUser("alice"); err != nil {
		t.Fatal(err)
	}
	client, err := os.ReadFile(filepath.Join(usersDir, "alice.conf"))
	if err != nil {
		t.Fatal(err)
	}
	serverData, err := os.ReadFile(filepath.Join(serverDir, "wg7.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(client), "Address = 10.7.0.2/32") {
		t.Fatalf("unexpected client config:\n%s", client)
	}
	if !strings.Contains(string(serverData), "# Name = alice") {
		t.Fatalf("unexpected server config:\n%s", serverData)
	}
	if got := fake.calls[len(fake.calls)-1]; got != "sudo systemctl restart wg-quick@wg7.service" {
		t.Fatalf("restart call = %q", got)
	}
}

func TestCustomConfigTemplatesAreUsed(t *testing.T) {
	configureTestDirectories(t)
	templateDir := filepath.Join(managerDir, templatesDirName)
	if err := os.MkdirAll(templateDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(serverTemplate.path(), []byte("server={{ .Alias }} users={{ len .Users }}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(clientTemplate.path(), []byte("client={{ .Name }} address={{ .ClientLocalAddress }}\n"), 0600); err != nil {
		t.Fatal(err)
	}

	serverConfig := validTestServerConfig()
	serverConfig.Users = []UserConfig{validTestUserConfig("alice", "10.7.0.2/32")}
	if err := writeServerConfig(serverConfig, serverConfig.Alias); err != nil {
		t.Fatal(err)
	}
	if err := writeClientConfig(serverConfig.Users[0], "alice"); err != nil {
		t.Fatal(err)
	}

	serverData, err := os.ReadFile(filepath.Join(serverDir, "wg7.conf"))
	if err != nil {
		t.Fatal(err)
	}
	clientData, err := os.ReadFile(filepath.Join(usersDir, "alice.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if string(serverData) != "server=wg7 users=1\n" {
		t.Fatalf("unexpected server configuration %q", serverData)
	}
	if string(clientData) != "client=alice address=10.7.0.2/32\n" {
		t.Fatalf("unexpected client configuration %q", clientData)
	}
}

func TestEnsureConfigTemplatesPreservesChanges(t *testing.T) {
	configureTestDirectories(t)
	if err := ensureConfigTemplates(); err != nil {
		t.Fatal(err)
	}
	custom := []byte("custom server template\n")
	if err := os.WriteFile(serverTemplate.path(), custom, 0600); err != nil {
		t.Fatal(err)
	}
	if err := ensureConfigTemplates(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(serverTemplate.path())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(custom) {
		t.Fatalf("server template was overwritten: %q", got)
	}
}

func TestWriteAndResetConfigTemplate(t *testing.T) {
	configureTestDirectories(t)
	custom := []byte("client={{ .Name }}\n")
	if err := WriteConfigTemplate("client", custom); err != nil {
		t.Fatal(err)
	}
	got, err := ReadConfigTemplate("client")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(custom) {
		t.Fatalf("client template = %q, want %q", got, custom)
	}
	if err := WriteConfigTemplate("client", []byte("{{")); err == nil {
		t.Fatal("expected invalid-template error")
	}
	if err := WriteConfigTemplate("client", nil); err == nil {
		t.Fatal("expected empty-template error")
	}
	got, err = ReadConfigTemplate("client")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(custom) {
		t.Fatalf("invalid update replaced client template: %q", got)
	}
	if err := ResetConfigTemplate("client"); err != nil {
		t.Fatal(err)
	}
	got, err = ReadConfigTemplate("client")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != defaultClientTemplate {
		t.Fatalf("client template was not reset: %q", got)
	}
}

func TestConfigTemplateRejectsUnknownName(t *testing.T) {
	configureTestDirectories(t)
	if _, err := ReadConfigTemplate("other"); err == nil {
		t.Fatal("expected unknown-template error")
	}
}

func TestInvalidCustomTemplateDoesNotReplaceConfiguration(t *testing.T) {
	configureTestDirectories(t)
	if err := ensureConfigTemplates(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(serverTemplate.path(), []byte("{{ .UnknownField }}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(serverDir, "wg7.conf")
	if err := os.WriteFile(path, []byte("existing\n"), 0600); err != nil {
		t.Fatal(err)
	}
	err := writeServerConfig(validTestServerConfig(), "wg7")
	if err == nil || !strings.Contains(err.Error(), "UnknownField") {
		t.Fatalf("expected missing-field error, got %v", err)
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "existing\n" {
		t.Fatalf("configuration changed after render error: %q", got)
	}
}

func TestPrepareSystemCreatesPrivateDirectoriesAndEnablesForwarding(t *testing.T) {
	configureTestDirectories(t)
	fake := &fakeRunner{}
	previous := commandRunner
	commandRunner = fake
	t.Cleanup(func() { commandRunner = previous })

	if err := prepareSystem(); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{serverDir, managerDir, userConfigDir, usersDir, tcConfigDir, filepath.Join(managerDir, templatesDirName)} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0700 {
			t.Fatalf("permissions for %s = %o, want 700", dir, got)
		}
	}
	for _, path := range []string{serverTemplate.path(), clientTemplate.path()} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0600 {
			t.Fatalf("permissions for %s = %o, want 600", path, got)
		}
	}
	data, err := os.ReadFile(sysctlFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "net.ipv4.ip_forward=1\n" {
		t.Fatalf("sysctl configuration = %q", data)
	}
	want := "sysctl -p " + sysctlFile
	if len(fake.calls) != 1 || fake.calls[0] != want {
		t.Fatalf("calls = %v, want %q", fake.calls, want)
	}
}

func configureTestDirectories(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	oldServerDir, oldManagerDir := serverDir, managerDir
	oldUserConfigDir, oldUsersDir := userConfigDir, usersDir
	oldTcConfigDir, oldSysctlFile := tcConfigDir, sysctlFile
	serverDir = filepath.Join(root, "wireguard")
	managerDir = filepath.Join(root, "manager")
	userConfigDir = filepath.Join(root, "configs")
	usersDir = filepath.Join(root, "users")
	tcConfigDir = filepath.Join(root, "tc")
	sysctlFile = filepath.Join(root, "90-gwg.conf")
	t.Cleanup(func() {
		serverDir, managerDir = oldServerDir, oldManagerDir
		userConfigDir, usersDir = oldUserConfigDir, oldUsersDir
		tcConfigDir, sysctlFile = oldTcConfigDir, oldSysctlFile
	})
	for _, dir := range []string{serverDir, managerDir, userConfigDir, usersDir, tcConfigDir} {
		if err := os.MkdirAll(dir, 0750); err != nil {
			t.Fatal(err)
		}
	}

}

func validTestServerConfig() WgServerConfig {
	return WgServerConfig{
		ServerPrivateKey: testKey,
		ServerPublicKey:  testKey,
		LocalAddress:     "10.7.0.1/24",
		PublicAddress:    "192.0.2.10",
		ListenPort:       51820,
		Eth:              "eth0",
		Alias:            "wg7",
		DnsResolv:        "8.8.8.8",
	}
}

func validTestUserConfig(name, address string) UserConfig {
	return UserConfig{
		ClientPrivateKey:   testKey,
		ClientPublicKey:    testKey,
		ClientLocalAddress: address,
		ServerPublicKey:    testKey,
		ServerIp:           "192.0.2.10",
		ServerPort:         51820,
		DnsResolv:          "8.8.8.8",
		Name:               name,
		Status:             "active",
	}
}
