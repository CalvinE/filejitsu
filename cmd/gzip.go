package cmd

import (
	"compress/gzip"
	"context"
	"errors"
	"log/slog"
	"time"

	fgzip "github.com/calvine/filejitsu/gzip"
	"github.com/spf13/cobra"
)

type GZipCompressionLevel string

type GZIPHeaderArgs struct {
	Comment string `json:"comment"`
	Extra   []byte `json:"extra"`
	Name    string `json:"name"`
	ModTime string `json:"modTime"`
	OS      int    `json:"os"`
}

type GZIPArgs struct {
	InputText        string               `json:"inputText"`
	CompressionLevel GZipCompressionLevel `json:"compressionLevel"`
	Header           GZIPHeaderArgs       `json:"gzipHeader"`
}

type GUNZIPArgs struct {
	InputText string `json:"inputText"`
}

var (
	errInvalidGZipLevel = errors.New("invalid GZIP level provided")

	gzipArgs = GZIPArgs{}

	gunzipArgs = GUNZIPArgs{}
)

const (
	gzipCommandName   string = "gzip"
	gunzipCommandName string = "gunzip"

	noCompression      GZipCompressionLevel = "NoCompression"
	bestSpeed          GZipCompressionLevel = "BestSpeed"
	bestCompression    GZipCompressionLevel = "BestCompression"
	huffmanOnly        GZipCompressionLevel = "HuffmanOnly"
	defaultCompression GZipCompressionLevel = "DefaultCompression"
)

func newGZIPCommand() *cobra.Command {
	return &cobra.Command{
		Use:     gzipCommandName,
		Aliases: []string{"gz"},
		Short:   "gzip the input provided",
		// TODO: write Long
		RunE: runGZIP,
	}
}

func newGUNZIPCommand() *cobra.Command {
	return &cobra.Command{
		Use:     gunzipCommandName,
		Aliases: []string{"guz"},
		Short:   "gunzip the input provided",
		// TODO: write long
		RunE: runGUNZIP,
	}
}

func GZipCompressionLevelToLevel(level GZipCompressionLevel) (int, error) {
	switch level {
	case noCompression:
		return gzip.NoCompression, nil
	case bestSpeed:
		return gzip.BestSpeed, nil
	case bestCompression:
		return gzip.BestCompression, nil
	case huffmanOnly:
		return gzip.HuffmanOnly, nil
	case defaultCompression:
		return gzip.DefaultCompression, nil
	}
	return 0, errInvalidGZipLevel
}

func gzipInit(parentCmd *cobra.Command) {
	gzipCommand := newGZIPCommand()
	gzipCommand.PersistentFlags().StringVarP(&gzipArgs.InputText, "inputText", "t", "", "Text to pass in as input")
	gzipCommand.PersistentFlags().StringVarP((*string)(&gzipArgs.CompressionLevel), "compressionLevel", "q", string(defaultCompression), "The compression level to use for gzip compression")
	gzipCommand.PersistentFlags().StringVarP(&gzipArgs.Header.Comment, "comment", "m", "", "The comment to place in the gzip header")
	gzipCommand.PersistentFlags().StringVarP(&gzipArgs.Header.Name, "name", "n", "", "The name to place in the gzip headers")
	gzipCommand.PersistentFlags().StringVar(&gzipArgs.Header.ModTime, "modTime", "", "the modTime to place in the gzip headers. Uses the format 2006-01-02 15:04:05")
	gzipCommand.PersistentFlags().BytesHexVarP(&gzipArgs.Header.Extra, "extra", "e", nil, "The extra data to place in the gzip header")
	parentCmd.AddCommand(gzipCommand)
	gunzipCommand := newGUNZIPCommand()
	gunzipCommand.PersistentFlags().StringVarP(&gunzipArgs.InputText, "inputText", "t", "", "Text to pass in as input")
	parentCmd.AddCommand(gunzipCommand)
}

func validateGZIPArgs(ctx context.Context, args GZIPArgs) (fgzip.CompressParams, error) {
	params := fgzip.CompressParams{}
	params.Input = getInputReader(commandLogger, inputFile, args.InputText)
	params.Header = gzip.Header{
		Comment: args.Header.Comment,
		Extra:   args.Header.Extra,
		Name:    args.Header.Name,
		OS:      byte(args.Header.OS),
	}
	if len(args.Header.ModTime) > 0 {
		modTime, err := time.Parse(time.DateTime, args.Header.ModTime)
		if err != nil {
			commandLogger.Error("failed to parse modTime", slog.String("timeString", args.Header.ModTime), slog.String("errorMessage", err.Error()))
			return params, err
		}
		params.Header.ModTime = modTime
	}
	params.Output = outputFile
	compressionLevel, err := GZipCompressionLevelToLevel(args.CompressionLevel)
	if err != nil {
		commandLogger.Error("invalid compression level provided", slog.String("compressionLevel", string(args.CompressionLevel)), slog.String("errorMessage", err.Error()))
		return params, err
	}
	params.Level = compressionLevel
	return params, nil
}

func runGZIP(cmd *cobra.Command, args []string) error {
	params, err := validateGZIPArgs(cmd.Context(), gzipArgs)
	if err != nil {
		commandLogger.Error("failed to validate gzip args", slog.String("errorMessage", err.Error()))
		return err
	}

	// TODO: populate the header with input file info if args are not set and input is a real file
	if err := fgzip.Compress(commandLogger, params); err != nil {
		commandLogger.Error("failed to gzip input", slog.String("errorMessage", err.Error()))
		return err
	}
	return nil
}

func validateGUNZIPArgs(cxt context.Context, args GUNZIPArgs) (fgzip.DecompressParams, error) {
	params := fgzip.DecompressParams{}
	params.Input = getInputReader(commandLogger, inputFile, args.InputText)
	params.Output = outputFile
	return params, nil
}

func runGUNZIP(cmd *cobra.Command, args []string) error {
	params, err := validateGUNZIPArgs(cmd.Context(), gunzipArgs)
	if err != nil {
		commandLogger.Error("failed to validate gunzip args", slog.String("errorMessage", err.Error()))
		return err
	}
	// TODO: what to do with header?
	// My thoughts are to have an init function that returns the header before decompressing.
	// that way you can use the header to prep a decompress target file if that is what someone desires.
	// For now I will ignore the header, until I find a need for it, or someone requests the functionality.
	header, err := fgzip.Decompress(commandLogger, params)
	if err != nil {
		commandLogger.Error("failed to gunzip input", slog.String("errorMessage", err.Error()))
		return err
	}
	commandLogger.Debug("gzip header retrieved", slog.Any("header", header))
	return nil
}
