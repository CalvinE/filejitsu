package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"log/slog"

	"github.com/calvine/filejitsu/encrypt"
	"github.com/spf13/cobra"
)

var encryptDecryptArgs = EncryptDecryptArgs{}

const (
	encryptCommandName     = "encrypt"
	decryptCommandName     = "decrypt"
	passthroughCommandName = "passthrough"
)

func newEncryptCommand() *cobra.Command {
	return &cobra.Command{
		Use:     encryptCommandName,
		Aliases: []string{"encr"},
		Short:   "encrypt data provided",
		Long:    "encrypt data provided using AES-256",
		RunE: func(cmd *cobra.Command, args []string) error {
			encryptDecryptArgs.Operation = encrypt.OpEncrypt
			return encryptDecryptRun(cmd, args)
		},
	}
}

func newDecryptCommand() *cobra.Command {
	return &cobra.Command{
		Use:     decryptCommandName,
		Aliases: []string{"dcry"},
		Short:   "decrypt data provided",
		Long:    "decrypt data provided using AES-256",
		RunE: func(cmd *cobra.Command, args []string) error {
			encryptDecryptArgs.Operation = encrypt.OpDecrypt
			return encryptDecryptRun(cmd, args)
		},
	}
}

func newPassthroughCommand() *cobra.Command {
	return &cobra.Command{
		Use:     passthroughCommandName,
		Aliases: []string{"pass"},
		Short:   "decrypt data provided",
		Long:    "decrypt data provided using AES-256",
		RunE: func(cmd *cobra.Command, args []string) error {
			encryptDecryptArgs.Operation = encrypt.OpPassThrough
			return encryptDecryptRun(cmd, args)
		},
	}
}

type EncryptDecryptArgs struct {
	InputText      string            `json:"inputText"`
	Passphrase     string            `json:"passphrase,omitempty"`
	PassphraseFile string            `json:"passphraseFile,omitempty"`
	Operation      encrypt.Operation `json:"operation"`
}

func validateEncryptArgs(ctx context.Context, args EncryptDecryptArgs) (encrypt.Params, error) {
	params := encrypt.Params{}
	if len(args.Passphrase) == 0 && len(args.PassphraseFile) == 0 {
		err := errors.New("passphrase or passphraseFile are required")
		return params, err
	}

	params.Input = getInputReader(commandLogger, inputFile, args.InputText)
	// if len(encryptDecryptArgs.InputText) > 0 {
	// 	commandLogger.Info("using provided text instead of input file")
	// 	params.Input = bytes.NewBufferString(encryptDecryptArgs.InputText)
	// } else {
	// 	commandLogger.Info("reading input from inputFile", slog.String("inputFilePath", inputPath))
	// 	params.Input = inputFile
	// }
	params.Output = outputFile

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

func encryptDecryptInit(parentCmd *cobra.Command) {
	encryptCommand := newEncryptCommand()
	encryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.InputText, "inputText", "t", "", "Text to pass in as input")
	encryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.Passphrase, "passphrase", "p", "", "The passphrase used to encrypt the data.")
	encryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.PassphraseFile, "passphraseFile", "f", "", "The file which will be read to get the passphrase used for encryption")
	parentCmd.AddCommand(encryptCommand)
	decryptCommand := newDecryptCommand()
	decryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.InputText, "inputText", "t", "", "Text to pass in as input")
	decryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.Passphrase, "passphrase", "p", "", "The passphrase used to encrypt the data.")
	decryptCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.PassphraseFile, "passphraseFile", "f", "", "The file which will be read to get the passphrase used for encryption")
	parentCmd.AddCommand(decryptCommand)
	passThroughCommand := newPassthroughCommand()
	passThroughCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.InputText, "inputText", "t", "", "Text to pass in as input")
	passThroughCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.Passphrase, "passphrase", "p", "", "The passphrase used to encrypt the data.")
	passThroughCommand.PersistentFlags().StringVarP(&encryptDecryptArgs.PassphraseFile, "passphraseFile", "f", "", "The file which will be read to get the passphrase used for encryption")
	parentCmd.AddCommand(passThroughCommand)
}

func encryptDecryptRun(cmd *cobra.Command, args []string) error {
	params, err := validateEncryptArgs(cmd.Context(), encryptDecryptArgs)
	if err != nil {
		commandLogger.Error("failed to validate args", slog.String("errorMessage", err.Error()))
		return err
	}
	switch encryptDecryptArgs.Operation {
	case encrypt.OpDecrypt:
		commandLogger.Debug("decrypt operation selected")
		if err := encrypt.Decrypt(commandLogger, params); err != nil {
			commandLogger.Error("failed to decrypt data", slog.String("errorMessage", err.Error()))
			return err
		}
	case encrypt.OpEncrypt:
		commandLogger.Debug("encrypt operation selected")
		if err := encrypt.Encrypt(commandLogger, params); err != nil {
			commandLogger.Error("failed to encrypt data", slog.String("errorMessage", err.Error()))
			return err
		}
	case encrypt.OpPassThrough:
		commandLogger.Debug("passthrough operation selected")
		if err := encrypt.Passthrough(commandLogger, params); err != nil {
			commandLogger.Error("failed to passthrough data", slog.String("errorMessage", err.Error()))
			return err
		}
	default:
		err = fmt.Errorf("got invalid operation code: %v", encryptDecryptArgs.Operation)
		commandLogger.Error("bad operation code encountered", slog.String("errorMessage", err.Error()))
		return err
	}
	return nil
}
