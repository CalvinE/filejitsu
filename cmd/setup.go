package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"log/slog"

	"github.com/calvine/filejitsu/util"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type readerCloserFunc func(r io.Reader) error
type writerCloserFunc func(w io.Writer) error

func noopCloseReader(r io.Reader) error {
	return nil
}

func noopCloseWriter(w io.Writer) error {
	return nil
}

// func closeWriter(logger *slog.Logger, w io.Writer, name, path string) error {
// 	c, ok := w.(io.Closer)
// 	if ok {
// 		closeOutputErr := c.Close()
// 		if closeOutputErr != nil {
// 			errMsg := fmt.Sprintf("failed to close %s file", name)
// 			commandLogger.Error(errMsg, slog.String("path", path), slog.String("errorMessage", closeOutputErr.Error()))
// 			closeOutputErr = fmt.Errorf("%s: %w", errMsg, closeOutputErr)
// 			return closeOutputErr
// 		}
// 	}
// 	return nil
// }

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
	//TODO: take this out of global scope
	logOutputFile io.Writer
	commandLogger *slog.Logger
	inputPath     string
	//TODO: take this out of global scope
	inputFile  io.Reader
	outputPath string
	//TODO: take this out of global scope
	outputFile *bufio.Writer // *os.File
	startTime  time.Time
)

func NewRootCMD() *cobra.Command {
	var logFileCloser writerCloserFunc
	var inputFileCloser readerCloserFunc
	var outputFileCloser writerCloserFunc
	return &cobra.Command{
		Use:   "filejitsu",
		Short: "A CLI tool for File System tools",
		Long:  "A CLI tool for File System tools",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			if logOutputPath != stdErrFileName {
				logFile, err := os.OpenFile(logOutputPath, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("failed to setup logging output file (%s): %w", logOutputPath, err)
				}
				logOutputFile = logFile
				logFileCloser = util.TryCloseWriter
			} else {
				logOutputFile = cmd.ErrOrStderr() // os.Stderr
				logFileCloser = noopCloseWriter
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
				inputFileCloser = util.TryCloseReader
			} else {
				commandLogger.Debug("using stdin as input")
				inputFile = cmd.InOrStdin() // os.Stdin
				inputFileCloser = noopCloseReader
			}

			commandLogger.Debug("setting up output", slog.String("outputPath", outputPath))
			var oFile io.Writer
			if outputPath != stdOutFileName {
				commandLogger.Info("outputFile set to something other than stdout", slog.String("outputPath", outputPath))
				f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					commandLogger.Error("failed to open output file", slog.String("outputPath", outputPath), slog.String("errorMessage", err.Error()))
					return err
				}
				oFile = f
				outputFileCloser = util.TryCloseWriter
			} else {
				commandLogger.Debug("using stdout as output")
				// changed output from file to io.Writer to support cmd.OutOrStdout for testing.
				oFile = cmd.OutOrStdout() // os.Stdout
				outputFileCloser = noopCloseWriter
			}

			outputFile = bufio.NewWriter(oFile)

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
			closeInputErr := inputFileCloser(inputFile)
			if closeInputErr != nil {
				errMsg := "failed to close input file"
				commandLogger.Error(errMsg, slog.String("path", inputPath), slog.String("errorMessage", closeInputErr.Error()))
				closeInputErr = fmt.Errorf("%s: %w", errMsg, closeInputErr)
				hadError = true
			}
			commandLogger.Debug("closing output",
				slog.String("outputPath", outputPath),
			)
			commandLogger.Debug("flushing output file")
			if err := outputFile.Flush(); err != nil {
				commandLogger.Warn("failed to flush output file", slog.String("errorMessage", closeInputErr.Error()))
				// TODO: error
			}
			closeOutputErr := outputFileCloser(outputFile)
			if closeOutputErr != nil {
				errMsg := "failed to close output file"
				commandLogger.Error(errMsg, slog.String("path", outputPath), slog.String("errorMessage", closeOutputErr.Error()))
				closeOutputErr = fmt.Errorf("%s: %w", errMsg, closeOutputErr)
				hadError = true
			}
			commandLogger.Debug("closing log output",
				slog.String("logOutputPath", logOutputPath),
			)
			closeLogOutputErr := logFileCloser(logOutputFile)
			if closeLogOutputErr != nil {
				errMsg := "failed to close log output file"
				commandLogger.Error(errMsg, slog.String("path", logOutputPath), slog.String("errorMessage", closeLogOutputErr.Error()))
				closeLogOutputErr = fmt.Errorf("%s: %w", errMsg, closeLogOutputErr)
				hadError = true
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

func setUpLogger(logLevelString string, logOutput io.Writer) *slog.Logger {
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

func SetupCommand(_buildHash, _buildDate string) *cobra.Command {
	buildHash = _buildHash
	buildDate = _buildDate
	rootCmd := NewRootCMD()
	rootCmd.PersistentFlags().StringVarP(&logLevelString, "logLevel", "l", "none", "The log level for the command. Supports error, warn, info, debug")
	rootCmd.PersistentFlags().StringVar(&logOutputPath, "logOutput", stdErrFileName, "Where to write the logs from the command to. Default is stderr")
	rootCmd.PersistentFlags().StringVarP(&inputPath, "input", "i", stdInFileName, "Where to read the input of the command (If there is any). Default is stdin")
	rootCmd.PersistentFlags().StringVarP(&outputPath, "output", "o", stdOutFileName, "Where to write the output of the command. Default is stdout")
	bulkRenameInit(rootCmd)
	encryptDecryptInit(rootCmd)
	base64CommandInit(rootCmd)
	spaceAnalyzerInit(rootCmd)
	gzipInit(rootCmd)
	tarInit(rootCmd)
	return rootCmd
}

func getInputReader(logger *slog.Logger, inputFile io.Reader, inputText string) io.Reader {
	if len(inputText) > 0 {
		logger.Info("using provided text instead of input file")
		return bytes.NewBufferString(inputText)
	}
	logger.Info("reading input from inputFile", slog.String("inputFilePath", inputPath))
	return inputFile
}

func getPassphrase(logger *slog.Logger, passphraseFile string, passphrase string) ([]byte, error) {
	if len(passphrase) > 0 {
		logger.Debug("passphrase provided so taking it")
		return []byte(passphrase), nil
	} else {
		logger.Debug("passphrase file provided", slog.String("file", passphraseFile))
		data, err := os.ReadFile(passphraseFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read passphraseFile: %w", err)
		}
		return data, nil
	}
}
