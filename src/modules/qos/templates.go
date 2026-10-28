package qos

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/PavelMilanov/go-wg-manager/internal/atomicfile"
	"github.com/PavelMilanov/go-wg-manager/internal/layout"
)

// ServiceTemplate is a deployment example, not an automatically installed unit.
// It intentionally has no ExecStartPre root deletion and no ExecStop shaping gap.
//
//go:embed templates/qos.service.tmpl
var ServiceTemplate string

//go:embed templates/channel.json.tmpl
var defaultConfigTemplate string

// RenderService prepares the module's qos.service unit for manual review.
// It does not write files, invoke systemctl or apply shaping.
func RenderService(c Config) ([]byte, error) {
	if _, err := c.Plan(); err != nil {
		return nil, err
	}
	return renderArtifact("qos.service", ServiceTemplate, c, DefaultConfigPath)
}

func renderArtifact(name, source string, c Config, configPath string) ([]byte, error) {
	tmpl, err := template.New(name).Option("missingkey=error").Parse(source)
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	data := struct {
		Config
		ConfigPath string
	}{c, configPath}
	if err := tmpl.Execute(&output, data); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

// Initialize generates module-owned files only. No Netlink or systemctl calls.
func Initialize(c Config) error { return initialize(layout.QoSDir, c) }

// Generate rerenders the unit from local .tmpl without changing channel.json.
func Generate() error { return generate(layout.QoSDir, nil) }

func initialize(root string, c Config) error {
	if _, err := c.Plan(); err != nil {
		return err
	}
	return generate(root, &c)
}

func generate(root string, initial *Config) error {
	m := manager{stateDir: filepath.Join(root, "state")}
	unlock, err := m.lock()
	if err != nil {
		return err
	}
	defer unlock()
	for _, dir := range []string{root, filepath.Join(root, "templates"), filepath.Join(root, "generated")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	for name, source := range map[string]string{"channel.json.tmpl": defaultConfigTemplate, "qos.service.tmpl": ServiceTemplate} {
		path := filepath.Join(root, "templates", name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := atomicfile.Write(path, []byte(source), 0600); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	configPath := filepath.Join(root, "channel.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if initial == nil {
			return fmt.Errorf("module config missing; run gwu qos init first")
		}
		source, err := os.ReadFile(filepath.Join(root, "templates", "channel.json.tmpl"))
		if err != nil {
			return err
		}
		data, err := renderArtifact("channel.json", string(source), *initial, configPath)
		if err != nil {
			return err
		}
		decoded, err := decodeConfig(data)
		if err != nil {
			return err
		}
		if decoded != *initial {
			return fmt.Errorf("config template changed requested interface/rate")
		}
		if err := atomicfile.Write(configPath, data, 0600); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	c, err := ReadConfig(configPath)
	if err != nil {
		return err
	}
	source, err := os.ReadFile(filepath.Join(root, "templates", "qos.service.tmpl"))
	if err != nil {
		return err
	}
	unit, err := renderArtifact("qos.service", string(source), c, configPath)
	if err != nil {
		return err
	}
	return atomicfile.Write(filepath.Join(root, "generated", "qos.service"), unit, 0644)
}
