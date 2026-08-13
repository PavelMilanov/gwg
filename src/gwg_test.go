package main

import (
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestRunRejectsUnsafeUserNameBeforeFilesystemAccess(t *testing.T) {
	err := run([]string{"add", "-name", "../../root"})
	if err == nil || !strings.Contains(err.Error(), "invalid user name") {
		t.Fatalf("error = %v", err)
	}
}

func TestRunHelpStopsCommand(t *testing.T) {
	err := run([]string{"add", "-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("error = %v, want flag.ErrHelp", err)
	}
}
