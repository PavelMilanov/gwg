package server

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/PavelMilanov/go-wg-manager/internal/atomicfile"
)

const templatesDirName = "templates"

//go:embed templates/server.conf.tmpl
var defaultServerTemplate string

//go:embed templates/client.conf.tmpl
var defaultClientTemplate string

type configTemplate struct {
	filename string
	content  string
}

var (
	serverTemplate = configTemplate{filename: "server.conf.tmpl", content: defaultServerTemplate}
	clientTemplate = configTemplate{filename: "client.conf.tmpl", content: defaultClientTemplate}
)

func (config configTemplate) path() string {
	return filepath.Join(managerDir, templatesDirName, config.filename)
}

func (config configTemplate) ensure() error {
	path := config.path()
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect template %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create template directory: %w", err)
	}
	if err := atomicfile.Write(path, []byte(config.content), 0600); err != nil {
		return fmt.Errorf("create template %s: %w", path, err)
	}
	return nil
}

func (config configTemplate) render(value any) ([]byte, error) {
	if err := config.ensure(); err != nil {
		return nil, err
	}
	path := config.path()
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", path, err)
	}
	templ, err := template.New(config.filename).Option("missingkey=error").Parse(string(source))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", path, err)
	}
	var output bytes.Buffer
	if err := templ.Execute(&output, value); err != nil {
		return nil, fmt.Errorf("render template %s: %w", path, err)
	}
	return output.Bytes(), nil
}

func ensureConfigTemplates() error {
	for _, config := range []configTemplate{serverTemplate, clientTemplate} {
		if err := config.ensure(); err != nil {
			return err
		}
	}
	return nil
}

func selectConfigTemplate(name string) (configTemplate, error) {
	switch name {
	case "server":
		return serverTemplate, nil
	case "client":
		return clientTemplate, nil
	default:
		return configTemplate{}, fmt.Errorf("unknown template %q: use server or client", name)
	}
}

func ReadConfigTemplate(name string) ([]byte, error) {
	config, err := selectConfigTemplate(name)
	if err != nil {
		return nil, err
	}
	if err := config.ensure(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(config.path())
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", config.path(), err)
	}
	return data, nil
}

func WriteConfigTemplate(name string, source []byte) error {
	config, err := selectConfigTemplate(name)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(source)) == 0 {
		return fmt.Errorf("template %s is empty", config.path())
	}
	if _, err := template.New(config.filename).Option("missingkey=error").Parse(string(source)); err != nil {
		return fmt.Errorf("parse template %s: %w", config.path(), err)
	}
	if err := os.MkdirAll(filepath.Dir(config.path()), 0700); err != nil {
		return fmt.Errorf("create template directory: %w", err)
	}
	if err := atomicfile.Write(config.path(), source, 0600); err != nil {
		return fmt.Errorf("write template %s: %w", config.path(), err)
	}
	return nil
}

func ResetConfigTemplate(name string) error {
	config, err := selectConfigTemplate(name)
	if err != nil {
		return err
	}
	return WriteConfigTemplate(name, []byte(config.content))
}
