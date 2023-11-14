package cmd

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

const (
	stdInFileName  = "stdin"
	stdOutFileName = "stdout"
	stdErrFileName = "stderr"
)

// TODO: move from global variables to a struct that could possibly be passed around...
var (
	buildHash string
	buildDate string

	logLevelString string
	logOutputPath  string
	logOutputFile  *os.File
	commandLogger  *slog.Logger
	inputPath      string
	inputFile      *os.File
	outputPath     string
	outputFile     *os.File
	startTime      time.Time
)

var rootCmd = &cobra.Command{
	Use:   "filejitsu",
	Short: "A CLI tool for File System tools",
	Long:  "A CLI tool for File System tools",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		startTime = time.Now()
		logOutputFile = os.Stderr
		if logOutputPath != stdErrFileName {
			logFile, err := os.OpenFile(logOutputPath, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("failed to setup logging output file (%s): %w", logOutputPath, err)
			}
			logOutputFile = logFile
		}
		logger := setUpLogger(logLevelString, logOutputFile)
		runID := uuid.New().String()
		commandName := getFullCommandName(cmd)
		commandLogger = logger.With(slog.String("command", commandName), slog.String("runID", runID))
		commandLogger.Info("starting filejitsu",
			slog.Time("startTime", startTime),
			slog.String("buildHash", buildHash),
			slog.String("buildDate", buildDate),
			slog.String("logOutputPath", logOutputPath),
			slog.String("inputPath", inputPath),
			slog.String("outputPath", outputPath),
		)
		commandLogger.Debug("setting up input", slog.String("inputPath", inputPath))
		if inputPath != stdInFileName {
			commandLogger.Info("inputFile set to something other than stdin", slog.String("inputPath", inputPath))
			f, err := os.OpenFile(inputPath, os.O_RDONLY, 0644)
			if err != nil {
				commandLogger.Error("failed to open input file", slog.String("inputPath", inputPath), slog.String("errorMessage", err.Error()))
				return err
			}
			inputFile = f
		} else {
			commandLogger.Debug("using stdin as input")
			inputFile = os.Stdin
		}

		commandLogger.Debug("setting up output", slog.String("outputPath", outputPath))
		if outputPath != stdOutFileName {
			commandLogger.Info("outputFile set to something other than stdout", slog.String("outputPath", outputPath))
			f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				commandLogger.Error("failed to open output file", slog.String("outputPath", outputPath), slog.String("errorMessage", err.Error()))
				return err
			}
			outputFile = f
		} else {
			commandLogger.Debug("using stdout as output")
			outputFile = os.Stdout
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		commandLogger.Debug("command post run", slog.Time("endTime", endTime), slog.String("duration", duration.String()))
		hadError := false
		commandLogger.Debug("closing input",
			slog.String("inputPath", inputPath),
		)
		closeInputErr := inputFile.Close()
		if closeInputErr != nil {
			hadError = true
			errMsg := "failed to close input file"
			commandLogger.Error(errMsg, slog.String("inputPath", inputPath), slog.String("errorMessage", closeInputErr.Error()))
			closeInputErr = fmt.Errorf("%s: %w", errMsg, closeInputErr)
		}
		commandLogger.Debug("closing output",
			slog.String("outputPath", outputPath),
		)
		closeOutputErr := outputFile.Close()
		if closeOutputErr != nil {
			hadError = true
			errMsg := "failed to close output file"
			commandLogger.Error(errMsg, slog.String("outputPath", outputPath), slog.String("errorMessage", closeOutputErr.Error()))
			closeOutputErr = fmt.Errorf("%s: %w", errMsg, closeInputErr)
		}
		commandLogger.Debug("closing log output",
			slog.String("logOutputPath", logOutputPath),
		)
		closeLogOutputErr := logOutputFile.Close()
		if closeLogOutputErr != nil {
			hadError = true
			errMsg := "failed to close log output file"
			// commandLogger.Error(errMsg, slog.String("logOutputPath", logOutputPath), slog.String("errorMessage", closeLogOutputErr.Error()))
			closeLogOutputErr = fmt.Errorf("%s: %w", errMsg, closeInputErr)
		}
		if hadError {
			err := errors.New("post run error(s)")
			if closeInputErr != nil {
				err = fmt.Errorf("%w: %w", err, closeInputErr)
			}
			if closeOutputErr != nil {
				err = fmt.Errorf("%w: %w", err, closeOutputErr)
			}
			if closeLogOutputErr != nil {
				err = fmt.Errorf("%w: %w", err, closeLogOutputErr)
			}
			return err
		}
		return nil
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

func setUpLogger(logLevelString string, logOutput *os.File) *slog.Logger {
	var level slog.Level
	logWriter := logOutput
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

func SetupCommand(_buildHash, _buildDate string) {
	buildHash = _buildHash
	buildDate = _buildDate
	rootCmd.PersistentFlags().StringVarP(&logLevelString, "logLevel", "l", "none", "The log level for the command. Supports error, warn, info, debug")
	rootCmd.PersistentFlags().StringVar(&logOutputPath, "logOutput", stdErrFileName, "Where to write the logs from the command to. Default is stderr")
	rootCmd.PersistentFlags().StringVarP(&inputPath, "input", "i", stdInFileName, "Where to read the input of the command (If there is any). Default is stdin")
	rootCmd.PersistentFlags().StringVarP(&outputPath, "output", "o", stdOutFileName, "Where to write the output of the command. Default is stdout")
	bulkRenameInit()
	encryptDecryptInit()
	base64CommandInit()
	spaceAnalyzerInit()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("failed to execute: %v", err)
		os.Exit(1)
	}
}
