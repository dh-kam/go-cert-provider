package main

import (
	"fmt"
	"os"

	"github.com/dh-kam/go-cert-provider/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
