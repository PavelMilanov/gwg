package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Runner isolates operating-system commands from domain logic and tests.
type Runner interface {
	Output(name string, args ...string) ([]byte, error)
	OutputWithInput(input []byte, name string, args ...string) ([]byte, error)
	Run(name string, args ...string) error
}

type ExecRunner struct{}

func (ExecRunner) Output(name string, args ...string) ([]byte, error) {
	return output(nil, name, args...)
}

func (ExecRunner) OutputWithInput(input []byte, name string, args ...string) ([]byte, error) {
	return output(bytes.NewReader(input), name, args...)
}

func output(stdin io.Reader, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, formatError(name, args, out, err)
	}
	return out, nil
}

func (ExecRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return formatError(name, args, nil, err)
	}
	return nil
}

func formatError(name string, args []string, output []byte, err error) error {
	command := strings.Join(append([]string{name}, args...), " ")
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("run %s: %w", command, err)
	}
	return fmt.Errorf("run %s: %w: %s", command, err, message)
}
