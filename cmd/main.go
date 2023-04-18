package main

import (
	"github.com/spf13/cobra"
)

// TODO: implement a logger, slog?

var rootCmd = &cobra.Command{
	Use:   "filejitsu",
	Short: "A CLI tool for File System tools",
	Long:  "A CLI tool for File System tools",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func main() {
	bulkRenameInit()
	rootCmd.Execute()
}
