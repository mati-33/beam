package main

import (
	"fmt"
	"os"

	"github.com/mati-33/beam/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "expected 'emit' or 'absorb' commands\n")
		os.Exit(1)
	}

	var err error

	switch os.Args[1] {
	case "emit":
		err = cli.Emit()
	case "absorb":
		err = cli.Absorb()
	case "help":
		cli.Help()
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "\nan error occured: %v\n", err)
		os.Exit(1)
	}
}
