package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: whoopctl <command>")
	}

	switch args[0] {
	case "version":
		fmt.Println("whoopctl dev")
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
