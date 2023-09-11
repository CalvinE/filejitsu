package cmd

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
	logWriter := os.Stderr
	switch strings.ToLower(logLevelString) {
	case "error":
		level = slog.LevelError.Level()
	case "warn":
		level = slog.LevelWarn.Level()
	case "info":
		level = slog.LevelInfo.Level()
	case "none":
		level = slog.LevelError.Level()
		var err error
		logWriter, err = os.Open(os.DevNull)
		if err != nil {
			panic("could not open os.DevNull!")
		}
	default:
		level = slog.LevelDebug.Level()
	}
	logger = *slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Just playing around with this
			return a
		},
	}))
}

func SetupCommand() {
	rootCmd.PersistentFlags().StringVarP(&logLevelString, "logLevel", "l", "none", "The log level for the command. Supports error, warn, info, debug")
	bulkRenameInit()
	encryptDecryptInit()
	base64CommandInit()
	spaceAnalyzerInit()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("failed to execute: %v", err)
		os.Exit(1)
	}
}
