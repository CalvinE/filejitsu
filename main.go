package main

import "github.com/calvine/filejitsu/cmd"

var (
	commitHash = "not_loaded_via_ld"
	buildDate  = "not_loaded_via_ld"
)

func main() {
	cmd.SetupCommand(commitHash, buildDate)
}
