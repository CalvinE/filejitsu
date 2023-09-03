package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/calvine/filejitsu/encrypt"
	"github.com/calvine/filejitsu/util"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

var encryptDecryptArgs = EncryptDecryptArgs{}

const (
	stdinName          = "_stdin"
	stdoutName         = "_stdout"
	encryptCommandName = "encrypt"
	decryptCommandName = "decrypt"
)

var encryptCommand = &cobra.Command{
	Use:     encryptCommandName,
	Aliases: []string{"encr"},
	Short:   "encrypt data provided",
	Long:    "encrypt data provided using AES-256",
	RunE: func(cmd *cobra.Command, args []string) error {
		encryptDecryptArgs.Operation = encrypt.OpEncrypt
		return Run(cmd, args)
	},
}

var decryptCommand = &cobra.Command{
	Use:     decryptCommandName,
	Aliases: []string{"dcry"},
	Short:   "decrypt data provided",
	Long:    "decrypt data provided using AES-256",
	RunE: func(cmd *cobra.Command, args []string) error {
		encryptDecryptArgs.Operation = encrypt.OpDecrypt
		return Run(cmd, args)
	},
}

type EncryptDecryptArgs struct {
	InputPath      string            `json:"filePath"`
	Passphrase     string            `json:"passphrase,omitempty"`
	PassphraseFile string            `json:"passphraseFile,omitempty"`
	OutputPath     string            `json:"outputFile"`
	Operation      encrypt.Operation `json:"operation"`
}

func validateEncryptArgs(ctx context.Context, args EncryptDecryptArgs) (encrypt.Params, error) {
	params := encrypt.Params{}
	if len(args.InputPath) == 0 {
		err := errors.New("inputPath is required")
		return params, err
	}
	if len(args.OutputPath) == 0 {
		err := errors.New("outputPath is required")
		return params, err
	}
	if len(args.Passphrase) == 0 && len(args.PassphraseFile) == 0 {
		err := errors.New("passphrase or passphraseFile are required")
		return params, err
	}
	if args.InputPath == stdinName {
		params.Input = os.Stdin
	} else {
		// open file reader
		f, err := util.OpenFile(args.InputPath, os.O_RDONLY, 0644)
		if err != nil {
			return params, fmt.Errorf("failed to open input file: %w", err)
		}
		params.Input = f
	}

	if args.OutputPath == stdoutName {
		params.Output = os.Stdout
	} else {
		// open file writer
		f, err := util.OpenFile(args.InputPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return params, fmt.Errorf("failed to open input file: %w", err)
		}
		params.Input = f
	}

	if len(args.Passphrase) > 0 {
		params.Passphrase = []byte(args.Passphrase)
	} else {
		data, err := os.ReadFile(args.PassphraseFile)
		if err != nil {
			return params, fmt.Errorf("failed to read passphraseFile: %w", err)
		}
		params.Passphrase = data
	}
	return params, nil
}

func encryptDecryptInit() {
	rootCmd.AddCommand(encryptCommand)
	var operation = "encrypt"
	encryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.InputPath, "input", "i", stdinName, fmt.Sprintf("The input to %s. Can be a file. If not specified it will use %s", operation, stdinName))
	encryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.OutputPath, "output", "o", stdoutName, fmt.Sprintf("The location where the %sed data will be placed. Can be a file. If not specified it will use %s", operation, stdoutName))
	encryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.Passphrase, "passphrase", "p", "", "The passphrase used to encrypt the data.")
	encryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.PassphraseFile, "passphraseFile", "f", "", "The file which will be read to get the passphrase used for encryption/")
	operation = "decrypt"
	decryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.InputPath, "input", "i", stdinName, fmt.Sprintf("The input to %s. Can be a file. If not specified it will use %s", operation, stdinName))
	decryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.OutputPath, "output", "o", stdoutName, fmt.Sprintf("The location where the %sed data will be placed. Can be a file. If not specified it will use %s", operation, stdoutName))
	decryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.Passphrase, "passphrase", "p", "", "The passphrase used to encrypt the data.")
	decryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.PassphraseFile, "passphraseFile", "f", "", "The file which will be read to get the passphrase used for encryption/")
	rootCmd.AddCommand(decryptCommand)
}

func attemptToCloseStreams(logger *slog.Logger, params encrypt.Params) error {
	logger.Debug("attempting to close input and output streams")
	readCloser, ok := params.Input.(io.ReadCloser)
	var err error = nil
	var readCloserErr error = nil
	if ok {
		err = readCloser.Close()
		readCloserErr = err
		if err != nil {
			logger.Error("failed to close input steam", slog.String("errorMessage", err.Error()))
		}
	} else {
		logger.Debug("input was not a read closer")
	}
	writeCloser, ok := params.Output.(io.WriteCloser)
	if ok {
		err = writeCloser.Close()
		if err != nil {
			logger.Error("failed to close output stream", slog.String("errorMessage", err.Error()))
			if readCloserErr != nil {
				err = fmt.Errorf("failed to close input and output: input - %w output - %w", readCloserErr, err)
			}
		}
	} else {
		logger.Debug("output was not a write closer")
	}
	return err
}

func Run(cmd *cobra.Command, args []string) error {
	commandLogger := logger.With(slog.String("commandName", encryptCommandName), slog.String("operation", encryptDecryptArgs.Operation.String()))
	commandLogger.Debug("starting command",
		slog.Any("args", encryptDecryptArgs),
	)
	defer commandLogger.Debug("ending command",
		slog.String("name", encryptCommandName),
	)
	params, err := validateEncryptArgs(cmd.Context(), encryptDecryptArgs)
	if err != nil {
		logger.Error("failed to validate args", slog.String("errorMessage", err.Error()))
		return err
	}
	switch encryptDecryptArgs.Operation {
	case encrypt.OpDecrypt:
		if err := encrypt.Decrypt(commandLogger, params); err != nil {
			commandLogger.Error("failed to decrypt file", slog.String("errorMessage", err.Error()))
			return err
		}
	case encrypt.OpEncrypt:
		if err := encrypt.Encrypt(commandLogger, params); err != nil {
			commandLogger.Error("failed to encrypt file", slog.String("errorMessage", err.Error()))
			return err
		}
	default:
		err = fmt.Errorf("got invalid operation code: %v", encryptDecryptArgs.Operation)
		logger.Error("bad operation code encountered", slog.String("errorMessage", err.Error()))
		return err
	}
	defer attemptToCloseStreams(commandLogger, params)
	return nil
}
