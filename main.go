package main

import (
	"fmt"
	"os"

	"github.com/calvine/filejitsu/cmd"
)

var (
	commitHash = "not_populated"
	buildDate  = "not_populated"
	buildTag   = "not_populated"
)

func main() {
	command := cmd.SetupCommand(commitHash, buildDate, buildTag)
	if err := command.Execute(); err != nil {
		fmt.Printf("failed to execute: %v", err)
		os.Exit(1)
	}
}
