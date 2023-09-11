package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

var commandLogger *slog.Logger

var logLevelString string

var startTime time.Time

var rootCmd = &cobra.Command{
	Use:   "filejitsu",
	Short: "A CLI tool for File System tools",
	Long:  "A CLI tool for File System tools",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		startTime = time.Now()
		logger := setUpLogger(logLevelString)
		runID := uuid.New().String()
		commandName := getFullCommandName(cmd)
		commandLogger = logger.With(slog.String("command", commandName), slog.String("runID", runID))
		commandLogger.Debug("command pre run", slog.Time("startTime", startTime))
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		commandLogger.Debug("command post run", slog.Time("endTime", endTime), slog.String("duration", duration.String()))
	},
	Run: func(cmd *cobra.Command, args []string) {},
}

func getFullCommandName(cmd *cobra.Command) string {
	commandParts := make([]string, 0)
	commandParts = append(commandParts, cmd.Name())
	parentCommand := cmd.Parent()
	for parentCommand != nil {
		commandParts = append(commandParts, parentCommand.Name())
		parentCommand = parentCommand.Parent()
	}
	slices.Reverse(commandParts)
	return strings.Join(commandParts, "->")
}

func setUpLogger(logLevelString string) *slog.Logger {
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
	logger := slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Just playing around with this
			return a
		},
	}))
	return logger
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
