package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

var rootCmd = &cobra.Command{
	Use:   "filejitsu",
	Short: "A CLI tool for File System tools",
	Long:  "A CLI tool for File System tools",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setUpLogger(logLevelString)
	},
	Run: func(cmd *cobra.Command, args []string) {},
}

var logger slog.Logger

var logLevelString string

func setUpLogger(logLevelString string) {
	var level slog.Level
	switch strings.ToLower(logLevelString) {
	case "error":
		level = slog.LevelError.Level()
	case "warn":
		level = slog.LevelWarn.Level()
	case "info":
		level = slog.LevelInfo.Level()
	default:
		level = slog.LevelDebug.Level()
	}
	logger = *slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Just playing around with this
			return a
		},
	}))
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&logLevelString, "logLevel", "l", "error", "The log level for the command. Supports error, warn, info, debug")
	bulkRenameInit()
	encryptDecryptInit()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("failed to execute: %v", err)
		os.Exit(1)
	}
}
