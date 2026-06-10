package main

import (
	"fmt"
	"os"

	"github.com/pod32g/omni-bucket/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
