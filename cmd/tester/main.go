package main

import (
	"os"

	"github.com/b-harvest/liquidity-stress-test/cmd/tester/cmd"
)

func main() {
	if err := cmd.RootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
