package cli

import (
	"context"
	"fmt"
	"io"
)

func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	_ = ctx
	_ = stderr

	if len(args) == 0 {
		return fmt.Errorf("usage: whoopctl <command>")
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, "whoopctl dev")
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
