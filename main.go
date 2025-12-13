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
	case "emit", "e":
		err = cli.Emit()
	case "absorb", "a":
		err = cli.Absorb()
	case "-h", "--help":
		cli.Help()
	case "-v", "--version":
		fmt.Println("beam version 0.0.1")
	default:
		cli.Usage()
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
