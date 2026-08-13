package command

import (
	"strings"
	"testing"
)

func TestExecRunnerOutputWithInput(t *testing.T) {
	runner := ExecRunner{}
	out, err := runner.OutputWithInput([]byte("payload"), "cat")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "payload" {
		t.Fatalf("output = %q", out)
	}
}

func TestExecRunnerIncludesStderrInError(t *testing.T) {
	runner := ExecRunner{}
	_, err := runner.Output("ls", "/path-that-does-not-exist-gwg")
	if err == nil || !strings.Contains(err.Error(), "path-that-does-not-exist-gwg") {
		t.Fatalf("error = %v", err)
	}
}
