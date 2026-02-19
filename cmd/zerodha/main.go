package main

import (
	"fmt"
	"os"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/cli"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitcode.Code(err))
	}
}
