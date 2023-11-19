package gzip

import (
	"compress/gzip"
	"io"
	"log/slog"
	"time"

	"github.com/calvine/filejitsu/util"
)

var emptyTime = time.Time{}

type Operation string

type Params struct {
	Input  io.Reader
	Output io.Writer
	Header gzip.Header
}

func Compress(logger *slog.Logger, params Params) error {
	out := gzip.NewWriter(params.Output)
	if len(params.Header.Comment) > 0 {
		out.Header.Comment = params.Header.Comment
	}
	if len(params.Header.Name) > 0 {
		out.Header.Name = params.Header.Name
	}
	if len(params.Header.Extra) > 0 {
		out.Header.Extra = params.Header.Extra
	}
	if params.Header.OS > 0 {
		out.Header.OS = params.Header.OS
	}
	if params.Header.ModTime != emptyTime {
		out.Header.ModTime = params.Header.ModTime
	}
	if err := util.ProcessStreams(logger, params.Input, out); err != nil {
		logger.Error("failed to write input to gzip output stream", slog.String("errorMessage", err.Error()))
		return err
	}
	if err := out.Flush(); err != nil {
		logger.Error("failed to flush gzip writer output", slog.String("errorMessage", err.Error()))
		return err
	}
	if err := out.Close(); err != nil {
		logger.Error("failed to close gzip writer", slog.String("errorMessage", err.Error()))
		return err
	}
	return nil
}

func Decompress(logger *slog.Logger, params Params) (gzip.Header, error) {
	in, err := gzip.NewReader(params.Input)
	if err != nil {
		logger.Error("failed to create gzip reader", slog.String("errorMessage", err.Error()))
		return in.Header, err
	}
	if err := util.ProcessStreams(logger, in, params.Output); err != nil {
		logger.Error("failed to read gzip data into output", slog.String("errorMessage", err.Error()))
		return in.Header, err
	}
	return in.Header, nil
}
