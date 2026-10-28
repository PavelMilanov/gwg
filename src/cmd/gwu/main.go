package main

import (
	"fmt"
	"os"

	"github.com/PavelMilanov/go-wg-manager/internal/gwucli"
)

func main() {
	if err := gwucli.NewCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
