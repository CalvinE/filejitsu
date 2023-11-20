package main

import (
	"fmt"
	"os"

	"github.com/calvine/filejitsu/cmd"
)

var (
	commitHash = "not_loaded_via_ld"
	buildDate  = "not_loaded_via_ld"
)

func main() {
	command := cmd.SetupCommand(commitHash, buildDate)
	if err := command.Execute(); err != nil {
		fmt.Printf("failed to execute: %v", err)
		os.Exit(1)
	}
}
