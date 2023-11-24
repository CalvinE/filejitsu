package cmd

import (
	"errors"
	"log/slog"

	"github.com/calvine/filejitsu/gzip"
	"github.com/calvine/filejitsu/tar"
	"github.com/calvine/filejitsu/util"
	"github.com/spf13/cobra"
)

// TODO: may be good to pass in tar file name for unpack so we can look for .tar.gz and automatically use gzip...
type TarOperation string

// TODO: change TarOperation it just a bool for pack or unpack
type TarPackageArgs struct {
	InputPath            string
	OutputPath           string
	Unpackage            bool
	UseGZip              bool
	GzipCompressionLevel gzip.GZipCompressionLevel
}

const (
	tarCommandName = "tar"
)

func newTarCommand() *cobra.Command {
	return &cobra.Command{
		Use:   tarCommandName,
		Short: "A tool for creating and unpacking tar archives",
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
	tarArgs = TarPackageArgs{}
)

func tarInit(parentCmd *cobra.Command) {
	tarCommand := newTarCommand()
	tarCommand.PersistentFlags().BoolVarP(&tarArgs.UseGZip, "useGzip", "z", false, "if present the contents being packaged or unpackaged will be gzipped")
	tarCommand.PersistentFlags().StringVarP((*string)(&tarArgs.GzipCompressionLevel), "compressionLevel", "q", string(gzip.DefaultCompression), "The compression level to use for gzip compression")
	tarCommand.PersistentFlags().BoolVarP(&tarArgs.Unpackage, "unpackage", "u", false, "if present the input tar package will be unpacked at the output path")
	parentCmd.AddCommand(tarCommand)
	util.HideGlobalFlags(tarCommand, []string{"input"})
}

func tarPackageRun(cmd *cobra.Command, args []string) error {
	// TODO: need validate args to check for if input path is specified for tar and not for untar (it would use -i)
	commandLogger.Debug("running tar package")
	if len(tarArgs.InputPath) == 0 {
		commandLogger.Debug("input path flag not set, trying to set from remaining args")
		numArgs := len(args)
		if numArgs == 1 {
			tarArgs.InputPath = args[0]
			commandLogger.Debug("pulling input path from remaining args")
		} else {
			errMsg := "number of arguments is not 1 so cannot pull input path from it"
			commandLogger.Error(errMsg, slog.Int("numArgs", numArgs))
			return errors.New(errMsg)
		}
	}
	commandLogger.Debug("input path set", slog.String("inputPath", tarArgs.InputPath))
	// TODO: have an issue where we need an input path for the tar command...
	// using the global input as currently designed will not work...
	// we need a path to talk and tar each file for packing...
	// Same likely true for output and unpacking...
	if err := tar.TarPackage(commandLogger, tar.TarPackageParams{
		InputPath: tarArgs.InputPath,
		Output:    outputFile,
		UseGzip:   tarArgs.UseGZip,
		GZIPOptions: tar.GZIPOptions{
			CompressionLevel: tarArgs.GzipCompressionLevel,
		},
	}); err != nil {
		commandLogger.Error("failed to package tar file", slog.String("errorMessage", err.Error()))
		return err
	}
	return nil
}

func tarUnpackageRun(cmd *cobra.Command, args []string) error {
	commandLogger.Debug("running tar unpackage")
	if len(tarArgs.OutputPath) == 0 {
		commandLogger.Debug("output path flag not set, trying to set from remaining args")
		numArgs := len(args)
		if numArgs == 1 {
			tarArgs.OutputPath = args[0]
			commandLogger.Debug("pulling output path from remaining args")
		} else {
			errMsg := "number of arguments is not 1 so cannot pull output path from it"
			commandLogger.Error(errMsg, slog.Int("numArgs", numArgs))
			return errors.New(errMsg)
		}
	}
	commandLogger.Debug("output path set", slog.String("outputPath", tarArgs.OutputPath))
	if err := tar.TarUnpackage(commandLogger, tar.TarUnpackageParams{
		Input:      inputFile,
		OutputPath: tarArgs.OutputPath,
		UseGzip:    tarArgs.UseGZip,
	}); err != nil {
		commandLogger.Error("failed to unpackage tar file", slog.String("errorMessage", err.Error()))
		return err
	}
	return nil
}
