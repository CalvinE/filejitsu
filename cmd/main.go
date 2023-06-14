package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

// TODO: implement a logger, slog?

var rootCmd = &cobra.Command{
	Use:   "filejitsu",
	Short: "A CLI tool for File System tools",
	Long:  "A CLI tool for File System tools",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var logger slog.Logger

func main() {
	logger = *slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug.Level(),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Just playing around with this
			return a
		},
	}))
	bulkRenameInit()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("failed to execute: %v", err)
		os.Exit(1)
	}
}
