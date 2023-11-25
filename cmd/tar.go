package cmd

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/calvine/filejitsu/gzip"
	"github.com/calvine/filejitsu/tar"
	"github.com/calvine/filejitsu/util"
	"github.com/spf13/cobra"
)

type TarArgs struct {
	InputPaths           []string
	OutputPath           string
	Unpackage            bool
	UseGZip              bool
	GzipCompressionLevel gzip.GZipCompressionLevel
	UseEncryption        bool
	Passphrase           string
	PassphraseFile       string
}

const (
	tarCommandName = "tar"
)

func newTarCommand() *cobra.Command {
	return &cobra.Command{
		Use:   tarCommandName,
		Short: "A tool for creating and unpacking tar archives",
		Long:  "A tool to package or unpackage a tar archive with optional gzip compression and AES256 encryption",
		RunE: func(cmd *cobra.Command, args []string) error {
			if tarArgs.Unpackage {
				return tarUnpackageRun(cmd, args)
			} else {
				return tarPackageRun(cmd, args)
			}
		},
	}
}

var (
	tarArgs = TarArgs{}
)

// TODO: add fuzzy match in input paths. * for instance... or regex?

func tarInit(parentCmd *cobra.Command) {
	tarCommand := newTarCommand()
	tarCommand.PersistentFlags().StringArrayVar(&tarArgs.InputPaths, "inputPath", nil, "The input path to tar. Can be file or directory. Can be specified multiple times - (USED ONLY WITH CREATING A TAR ARCHIVE I.E. NO unpackage flag)")
	tarCommand.PersistentFlags().StringVar(&tarArgs.OutputPath, "outputPath", "", "The output path to untar the contents of a tar archive to. Must be a directory - (USED ONLY WITH THE unpackage FLAG)")
	tarCommand.PersistentFlags().BoolVarP(&tarArgs.UseGZip, "useGzip", "z", false, "If present the contents being packaged will be gzipped or unpackaged will be gunzipped")
	tarCommand.PersistentFlags().StringVarP((*string)(&tarArgs.GzipCompressionLevel), "CompressionLevel", "q", string(gzip.DefaultCompression), "The compression level to use for gzip compression")
	tarCommand.PersistentFlags().BoolVarP(&tarArgs.Unpackage, "unpackage", "u", false, "If present the input tar package will be unpacked at the outputPath")
	tarCommand.PersistentFlags().BoolVarP(&tarArgs.UseEncryption, "encrypt", "e", false, "If present the tar will be encrypted while created, or decrypted while unpacked. Requires a passphrase or passphrase file be provided")
	tarCommand.PersistentFlags().StringVarP(&tarArgs.Passphrase, "passphrase", "p", "", "The passphrase used to encrypt or decrypt the data")
	tarCommand.PersistentFlags().StringVarP(&tarArgs.PassphraseFile, "passphraseFile", "f", "", "The file which will be read to get the passphrase used for encryption or decryption")
	parentCmd.AddCommand(tarCommand)
	util.HideGlobalFlags(tarCommand, map[string]util.FlagModifier{
		"input": {
			Hide:         false,
			UsagePostFix: "(NOT USED FOR TAR PACKAGING, USE inputPath FLAG INSTEAD)",
		},
		"output": {
			Hide:         false,
			UsagePostFix: "(NOT USED FOR TAR UNPACKING, USE outputPath FLAG INSTEAD)",
		},
	})
}

func ValidateTarPackageArgs(logger *slog.Logger, tarArgs TarArgs, args []string) (tar.TarPackageParams, error) {
	params := tar.TarPackageParams{}
	if tarArgs.Unpackage {
		return params, errors.New("unpackage flag set for package command")
	}
	if len(tarArgs.InputPaths) == 0 {
		logger.Debug("input path flag not set, trying to set from remaining args")
		numArgs := len(args)
		if numArgs > 0 {
			tarArgs.InputPaths = args
			logger.Debug("pulling input path from remaining args")
		} else {
			errMsg := "no arguments provided and inputPaths not set"
			logger.Error(errMsg, slog.Int("numArgs", numArgs))
			return params, errors.New(errMsg)
		}
	}
	params.InputPaths = tarArgs.InputPaths
	logger.Debug("input path set", slog.Any("inputPath", params.InputPaths))
	// gzip stuff
	if tarArgs.UseGZip {
		params.UseGzip = true
		params.GZIPOptions.CompressionLevel = tarArgs.GzipCompressionLevel
	}
	if tarArgs.UseEncryption {
		params.UseEncryption = true
		passphrase, err := getPassphrase(logger, tarArgs.PassphraseFile, tarArgs.Passphrase)
		if err != nil {
			errMsg := "error getting passphrase"
			logger.Error(errMsg, slog.String("errorMessage", err.Error()))
			return params, fmt.Errorf("%s: %w", errMsg, err)
		}
		params.EncryptionOptions.Passphrase = passphrase
	}
	params.Output = outputFile
	return params, nil
}

func tarPackageRun(cmd *cobra.Command, args []string) error {
	commandLogger.Debug("running tar package")
	params, err := ValidateTarPackageArgs(commandLogger, tarArgs, args)
	if err != nil {
		errMsg := "failed to validate tar packaging args"
		commandLogger.Error(errMsg, slog.String("errorMessage", err.Error()))
		return fmt.Errorf("%s: %w", errMsg, err)
	}
	if err := tar.TarPackage(commandLogger, params); err != nil {
		commandLogger.Error("failed to package tar file", slog.String("errorMessage", err.Error()))
		return err
	}
	return nil
}

func ValidateTarUnpackageArgs(logger *slog.Logger, tarArgs TarArgs, args []string) (tar.TarUnpackageParams, error) {
	params := tar.TarUnpackageParams{}
	if !tarArgs.Unpackage {
		return params, errors.New("unpackage flag not set for unpackage command")
	}
	params.Input = inputFile
	if len(tarArgs.OutputPath) == 0 {
		logger.Debug("output path flag not set, trying to set from remaining args")
		numArgs := len(args)
		if numArgs == 1 {
			tarArgs.OutputPath = args[0]
			logger.Debug("pulling output path from remaining args")
		} else {
			errMsg := "no arguments or too many arguments provided and output path not set"
			logger.Error(errMsg, slog.Int("numArgs", numArgs))
			return params, errors.New(errMsg)
		}
	}
	logger.Debug("setting outputPath", slog.String("outputPath", tarArgs.OutputPath))
	params.OutputPath = tarArgs.OutputPath
	// gzip stuff
	if tarArgs.UseGZip {
		params.UseGzip = true
	}
	if tarArgs.UseEncryption {
		params.UseEncryption = true
		passphrase, err := getPassphrase(logger, tarArgs.PassphraseFile, tarArgs.Passphrase)
		if err != nil {
			errMsg := "error getting passphrase"
			logger.Error(errMsg, slog.String("errorMessage", err.Error()))
			return params, fmt.Errorf("%s: %w", errMsg, err)
		}
		params.EncryptionOptions.Passphrase = passphrase
	}
	return params, nil
}

func tarUnpackageRun(cmd *cobra.Command, args []string) error {
	commandLogger.Debug("running tar unpackage")
	params, err := ValidateTarUnpackageArgs(commandLogger, tarArgs, args)
	if err != nil {
		errMsg := "un tar arg validation failed"
		commandLogger.Error(errMsg, slog.String("errorMessage", err.Error()))
		return fmt.Errorf("%s: %w", errMsg, err)
	}
	commandLogger.Debug("output path set", slog.String("outputPath", tarArgs.OutputPath))
	if err := tar.TarUnpackage(commandLogger, params); err != nil {
		commandLogger.Error("failed to unpackage tar file", slog.String("errorMessage", err.Error()))
		return err
	}
	return nil
}
