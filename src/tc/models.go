package tc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/PavelMilanov/go-wg-manager/internal/atomicfile"
	"github.com/PavelMilanov/go-wg-manager/internal/command"
	"github.com/PavelMilanov/go-wg-manager/paths"
)

var (
	tcDir                        = paths.TC_DIR
	serviceDir                   = "/etc/systemd/system"
	commandRunner command.Runner = command.ExecRunner{}
)

type TcConfig struct {
	Intf      string
	Speed     string
	FullSpeed string
	Classes   []TcClass
	Filters   []TcFilter
}

type TcClass struct {
	Class       string
	Description string
	CeilSpeed   string
	MinSpeed    string
}

type TcFilter struct {
	Description string
	UserIp      string
	Class       string
}

func writeJSONFile(name string, value any) error {
	data, err := json.MarshalIndent(value, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", name, err)
	}
	return atomicfile.Write(filepath.Join(tcDir, name), data, 0600)
}

func (tc *TcConfig) config() error {
	return writeJSONFile(paths.TC_FILE, tc)
}

func (tc *TcConfig) generate() error {
	templ, err := template.New("tc").Parse(TC_TEMPLATE)
	if err != nil {
		return fmt.Errorf("parse tc template: %w", err)
	}
	var output bytes.Buffer
	if err := templ.Execute(&output, tc); err != nil {
		return fmt.Errorf("render tc script: %w", err)
	}
	return atomicfile.Write(filepath.Join(tcDir, paths.TC_CONFIG_FILE), output.Bytes(), 0700)
}

func (tc *TcConfig) createService() error {
	templ, err := template.New("tc-service").Parse(TC_SERVICE_TEMPLATE)
	if err != nil {
		return fmt.Errorf("parse tc service template: %w", err)
	}
	var output bytes.Buffer
	if err := templ.Execute(&output, tc); err != nil {
		return fmt.Errorf("render tc service: %w", err)
	}
	source := filepath.Join(tcDir, paths.TC_SERVICE_FILE)
	if err := atomicfile.Write(source, output.Bytes(), 0644); err != nil {
		return err
	}
	destination := filepath.Join(serviceDir, paths.TC_SERVICE_FILE)
	if err := commandRunner.Run("sudo", "install", "-m", "0644", source, destination); err != nil {
		return err
	}
	if err := commandRunner.Run("sudo", "systemctl", "daemon-reload"); err != nil {
		return err
	}
	if err := commandRunner.Run("sudo", "systemctl", "enable", paths.TC_SERVICE_FILE); err != nil {
		return err
	}
	return commandRunner.Run("sudo", "systemctl", "start", paths.TC_SERVICE_FILE)
}

func (tc *TcConfig) removeService() error {
	disableErr := commandRunner.Run("sudo", "systemctl", "disable", "--now", paths.TC_SERVICE_FILE)
	removeErr := commandRunner.Run("sudo", "rm", "-f", filepath.Join(serviceDir, paths.TC_SERVICE_FILE))
	reloadErr := commandRunner.Run("sudo", "systemctl", "daemon-reload")
	return errors.Join(disableErr, removeErr, reloadErr)
}

func (tc *TcConfig) start() error {
	return commandRunner.Run("sudo", filepath.Join(tcDir, paths.TC_CONFIG_FILE))
}

func (tc *TcConfig) down() error {
	if err := validateInterfaceName(tc.Intf); err != nil {
		return err
	}
	return commandRunner.Run("sudo", "tc", "qdisc", "del", "dev", tc.Intf, "root")
}

func ensureTcDir() error {
	if err := os.MkdirAll(tcDir, 0750); err != nil {
		return fmt.Errorf("create tc directory: %w", err)
	}
	return nil
}
