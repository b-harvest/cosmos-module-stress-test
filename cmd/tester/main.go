package main

import (
	"os"

	"github.com/nodebreaker0-0/cosmos-module-stress-test/cmd/tester/cmd"
)

func main() {
	if err := cmd.RootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
