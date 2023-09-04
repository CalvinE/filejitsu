package cmd

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"

	"github.com/calvine/filejitsu/util"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

type Base64Args struct {
	Input string `json:"input"`
	//Output string `json:"output"`
	Decode            bool `json:"decode"`
	UseURLEncoding    bool `json:"useURLEncoding"`
	OmitPadding       bool `json:"omitPadding"`
	OmitEndingNewLine bool `json:"omitEndingNewLine"`
}

const base64CommandName = "base64"

var base64Command = &cobra.Command{
	Use:     base64CommandName,
	Aliases: []string{"b64"},
	Short:   "Base 64 encode / decode input",
	Long:    "Base 64 encode / decode input",
	RunE:    base64Run,
}

var base64Args = Base64Args{}

func base64CommandInit() {
	base64Command.PersistentFlags().StringVarP(&base64Args.Input, "input", "i", stdinName, "The input to base 64 encode.  Default is stdin.")
	//base64Command.PersistentFlags().StringVarP(&base64Args.Input, "output", "o", stdoutName, "Where to output the base 64 encoded input..  Default is stdout.")
	base64Command.PersistentFlags().BoolVarP(&base64Args.Decode, "decode", "d", false, "if provided the command will decode the input. By default this command encodes the input.")
	base64Command.PersistentFlags().BoolVarP(&base64Args.UseURLEncoding, "useUrlEncoding", "u", false, "if provided the command will encode / decode using url encoding. By default this command encodes / decodes using std encoding.")
	base64Command.PersistentFlags().BoolVarP(&base64Args.OmitPadding, "omitPadding", "n", false, "if provided the command will encode /decode the input with padding. By default this command encodes / decodes with padding.")
	base64Command.PersistentFlags().BoolVarP(&base64Args.OmitEndingNewLine, "omitEndingNewLine", "e", false, "if provided the command omit the newline at the end of the output.")
	rootCmd.AddCommand(base64Command)
}

func base64Run(cmd *cobra.Command, args []string) error {
	commandLogger := logger.With(slog.String("commandName", base64CommandName))
	commandLogger.Debug("starting command",
		slog.Any("args", base64Args),
	)
	defer commandLogger.Debug("ending command")
	var input io.Reader
	if base64Args.Input == stdinName {
		input = os.Stdin
	} else {
		input = bytes.NewBufferString(base64Args.Input)
	}
	defer util.AttemptToClose(commandLogger, input)
	targetEncoding := getBase64Encoding(commandLogger, base64Args.UseURLEncoding, base64Args.OmitPadding)
	output := os.Stdout
	if base64Args.Decode {
		// decode
		err := base64Decode(commandLogger, targetEncoding, input, output)
		if err != nil {
			commandLogger.Error("failed to base64 decode input to output", slog.String("errorMessage", err.Error()))
			return err
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
