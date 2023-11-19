package cmd

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"log/slog"

	"github.com/calvine/filejitsu/util"
	"github.com/spf13/cobra"
)

type base64EncodingStrategy struct {
	Name     string
	Encoding *base64.Encoding
}

type Base64Args struct {
	InputText         string `json:"inputText"`
	Decode            bool   `json:"decode"`
	UseURLEncoding    bool   `json:"useURLEncoding"`
	OmitPadding       bool   `json:"omitPadding"`
	OmitEndingNewLine bool   `json:"omitEndingNewLine"`
	IsB64DCommand     bool   `json:"isB64DCommand"`
}

var (
	base64EncodingStrategies = []base64EncodingStrategy{
		{
			Name:     "RawURLEncoding",
			Encoding: base64.RawURLEncoding,
		},
		{
			Name:     "URLEncoding",
			Encoding: base64.URLEncoding,
		},
		{
			Name:     "RawStdEncoding",
			Encoding: base64.RawStdEncoding,
		},
		{
			Name:     "StdEncoding",
			Encoding: base64.StdEncoding,
		},
	}
)

const (
	base64CommandName       = "base64"
	base64EncodeCommandName = "encode"
	base64DecodeCommandName = "decode"
)

var base64Command = &cobra.Command{
	Use:     base64CommandName,
	Aliases: []string{"b64"},
	Short:   "Base 64 encode / decode input",
	Long:    "Base 64 encode / decode input",
	RunE:    base64Run,
}

var base64EncodeCommand = &cobra.Command{
	Use:     base64EncodeCommandName,
	Aliases: []string{"e"},
	Short:   "Base 64 encode input",
	Long:    "Base 64 encode input",
	RunE: func(cmd *cobra.Command, args []string) error {
		base64Args.Decode = false
		return base64Run(cmd, args)
	},
}

var base64DecodeCommand = &cobra.Command{
	Use:     base64DecodeCommandName,
	Aliases: []string{"d"},
	Short:   "Base 64 decode input",
	Long:    "A simplified version of b64 -d. This command will try all encodings and succeed if any are successful",
	RunE: func(cmd *cobra.Command, args []string) error {
		base64Args.Decode = true
		base64Args.IsB64DCommand = true
		return base64Run(cmd, args)
	},
}

var base64Args = Base64Args{}

func base64CommandInit() {
	base64Command.PersistentFlags().StringVarP(&base64Args.InputText, "inputText", "t", "", "Text to pass in as input")
	base64Command.PersistentFlags().BoolVarP(&base64Args.Decode, "decode", "d", false, "if provided the command will decode the input. By default this command encodes the input.")
	base64Command.PersistentFlags().BoolVarP(&base64Args.UseURLEncoding, "useUrlEncoding", "u", false, "if provided the command will encode / decode using url encoding. By default this command encodes / decodes using std encoding.")
	base64Command.PersistentFlags().BoolVarP(&base64Args.OmitPadding, "omitPadding", "n", false, "if provided the command will encode /decode the input with padding. By default this command encodes / decodes with padding.")
	base64Command.PersistentFlags().BoolVarP(&base64Args.OmitEndingNewLine, "omitEndingNewLine", "e", false, "if provided the command omit the newline at the end of the output.")
	rootCmd.AddCommand(base64Command)
	base64EncodeCommand.PersistentFlags().StringVarP(&base64Args.InputText, "inputText", "t", "", "Text to pass in as input")
	base64EncodeCommand.PersistentFlags().BoolVarP(&base64Args.UseURLEncoding, "useUrlEncoding", "u", false, "if provided the command will encode / decode using url encoding. By default this command encodes / decodes using std encoding.")
	base64EncodeCommand.PersistentFlags().BoolVarP(&base64Args.OmitPadding, "omitPadding", "n", false, "if provided the command will encode /decode the input with padding. By default this command encodes / decodes with padding.")
	base64EncodeCommand.PersistentFlags().BoolVarP(&base64Args.OmitEndingNewLine, "omitEndingNewLine", "e", false, "if provided the command omit the newline at the end of the output.")
	base64Command.AddCommand(base64EncodeCommand)
	base64DecodeCommand.PersistentFlags().StringVarP(&base64Args.InputText, "inputText", "t", "", "Text to pass in as input")
	base64DecodeCommand.PersistentFlags().BoolVarP(&base64Args.OmitEndingNewLine, "omitEndingNewLine", "e", false, "if provided the command omit the newline at the end of the output.")
	base64Command.AddCommand(base64DecodeCommand)
}

func base64Run(cmd *cobra.Command, args []string) error {
	var input io.Reader
	if len(base64Args.InputText) > 0 {
		commandLogger.Info("using provided text instead of input file")
		input = bytes.NewBufferString(base64Args.InputText)
	} else {
		commandLogger.Info("reading input from inputFile", slog.String("inputFilePath", inputPath))
		input = inputFile
	}
	targetEncoding := getBase64Encoding(commandLogger, base64Args.UseURLEncoding, base64Args.OmitPadding)
	output := outputFile
	if base64Args.Decode {
		// decode
		if base64Args.IsB64DCommand {
			commandLogger.Debug("running simplified base64 decode command")
			err := robustBase64Decode(commandLogger, input, output)
			if err != nil {
				commandLogger.Error("failed to base64 decode input to output", slog.String("errorMessage", err.Error()))
				return err
			}
		} else {
			commandLogger.Debug("running normal base64 decode command")
			err := base64Decode(commandLogger, targetEncoding, input, output)
			if err != nil {
				commandLogger.Error("failed to base64 decode input to output", slog.String("errorMessage", err.Error()))
				return err
			}
		}
	} else {
		// encode
		err := base64Encode(commandLogger, targetEncoding, input, output)
		if err != nil {
			commandLogger.Error("failed to base64 encode input to output", slog.String("errorMessage", err.Error()))
			return err
		}
	}
	if !base64Args.OmitEndingNewLine {
		commandLogger.Debug("writing new line to end of output")
		_, err := output.Write([]byte(util.NewLine))
		if err != nil {
			commandLogger.Error("failed to write new line at end of output", slog.String("errorMessage", err.Error()))
			return err
		}
		return nil
	}
	commandLogger.Debug("skipping writing new line to output.")
	return nil
}

func robustBase64Decode(commandLogger *slog.Logger, input io.Reader, output io.Writer) error {
	commandLogger.Debug("decode operation selected")
	rawData, err := util.ReaderReadAll(commandLogger, input)
	if err != nil {
		errMsg := "failed to read all data from the input"
		commandLogger.Error(errMsg, slog.String("errorMessage", err.Error()))
		return fmt.Errorf("%s: %w", errMsg, err)
	}
	outerErr := errors.New("failed to base64 decode input")
	for _, e := range base64EncodingStrategies {
		inputBuffer := bytes.NewBuffer(rawData)
		commandLogger.Debug("attempting to base64 decode", slog.String("encodingUsed", e.Name))
		base64Decoder := base64.NewDecoder(e.Encoding, inputBuffer)
		// TODO: buffer the reads to the writer to avoid reading all of the data at once incase its a LOT of data? Perhaps only on the non robust decode? The try / fail approach here is a potential issue?
		data, err := util.ReaderReadAll(commandLogger, base64Decoder)
		if err != nil {
			commandLogger.Debug("failed to read data from base 64 decoder", slog.String("errorMessage", err.Error()), slog.String("encodingUsed", e.Name))
			outerErr = fmt.Errorf("%w (%s): %w", outerErr, e.Name, err)
			inputBuffer.UnreadByte()
			continue
		}
		bytesWritten, err := output.Write(data)
		if err != nil {
			errMsg := "failed to write decoded data to output"
			commandLogger.Debug(errMsg, slog.String("errorMessage", err.Error()), slog.String("encodingUsed", e.Name))
			return fmt.Errorf("%s (%s): %w", errMsg, e.Name, err)
		}
		commandLogger.Debug("wrote decoded input to output", slog.Int("bytesWritten", bytesWritten))
		return nil
	}
	return outerErr
}

func base64Decode(commandLogger *slog.Logger, targetEncoding *base64.Encoding, input io.Reader, output io.Writer) error {
	commandLogger.Debug("decode operation selected")
	base64Decoder := base64.NewDecoder(targetEncoding, input)
	data, err := util.ReaderReadAll(commandLogger, base64Decoder)
	if err != nil {
		commandLogger.Error("failed to read data from base 64 decoder", slog.String("errorMessage", err.Error()))
		return err
	}
	bytesWritten, err := output.Write(data)
	if err != nil {
		commandLogger.Error("failed to write decoded data to output", slog.String("errorMessage", err.Error()))
		return err
	}
	commandLogger.Debug("wrote decoded input to output", slog.Int("bytesWritten", bytesWritten))
	return nil
}

func base64Encode(commandLogger *slog.Logger, targetEncoding *base64.Encoding, input io.Reader, output io.Writer) error {
	commandLogger.Debug("encode operation selected")
	base64Encoder := base64.NewEncoder(targetEncoding, output)
	inputData, err := util.ReaderReadAll(commandLogger, input)
	if err != nil {
		commandLogger.Error("failed to read all data from input", slog.String("errorMessage", err.Error()))
		return err
	}
	bytesWritten, err := base64Encoder.Write(inputData)
	if err != nil {
		commandLogger.Error("failed to write input to output", slog.String("errorMessage", err.Error()))
		return err
	}
	commandLogger.Debug("write input to output", slog.Int("bytesWritten", bytesWritten))
	err = base64Encoder.Close()
	if err != nil {
		commandLogger.Error("failed to close output", slog.String("errorMessage", err.Error()))
		return err
	}
	return nil
}

func getBase64Encoding(logger *slog.Logger, useURLEncoding, omitPadding bool) *base64.Encoding {
	if useURLEncoding {
		if omitPadding {
			logger.Debug("selected base64.RawURLEncoding")
			return base64.RawURLEncoding
		}
		logger.Debug("selected base64.URLEncoding")
		return base64.URLEncoding
	}
	if omitPadding {
		logger.Debug("selected base64.RawStdEncoding")
		return base64.RawStdEncoding
	}
	logger.Debug("selected base64.StdEncoding")
	return base64.StdEncoding
}
