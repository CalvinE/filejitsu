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

type CompressParams struct {
	Level  int
	Header gzip.Header
	Input  io.Reader
	Output io.Writer
}

type DecompressParams struct {
	Input  io.Reader
	Output io.Writer
}

func Compress(logger *slog.Logger, params CompressParams) error {
	logger.Debug("creating gzip writer")
	out, err := gzip.NewWriterLevel(params.Output, params.Level)
	if err != nil {
		logger.Debug("failed to create new gzip writer with compression level", slog.Int("compressionLevel", params.Level), slog.String("errorMessage", err.Error()))
		return err
	}
	if len(params.Header.Comment) > 0 {
		logger.Debug("adding header comment", slog.String("comment", params.Header.Comment))
		out.Header.Comment = params.Header.Comment
	}
	if len(params.Header.Name) > 0 {
		logger.Debug("adding header name", slog.String("name", params.Header.Name))
		out.Header.Name = params.Header.Name
	}
	if len(params.Header.Extra) > 0 {
		logger.Debug("adding header extra data", slog.Any("extraData", params.Header.Extra))
		out.Header.Extra = params.Header.Extra
	}
	if params.Header.OS > 0 {
		logger.Debug("adding header os", slog.Int("os", int(params.Header.OS)))
		out.Header.OS = params.Header.OS
	}
	if params.Header.ModTime != emptyTime {
		slog.Debug("adding header modTime", slog.Time("modTime", params.Header.ModTime))
		out.Header.ModTime = params.Header.ModTime
	}
	logger.Debug("writing input to gzip writer")
	if err := util.ProcessStreams(logger, params.Input, out); err != nil {
		logger.Error("failed to write input to gzip output stream", slog.String("errorMessage", err.Error()))
		return err
	}
	logger.Debug("closing gzip writer")
	if err := out.Flush(); err != nil {
		logger.Error("failed to flush gzip writer", slog.String("errorMessage", err.Error()))
		return err
	}
	if err := out.Close(); err != nil {
		logger.Error("failed to close gzip writer", slog.String("errorMessage", err.Error()))
		return err
	}
	return nil
}

func Decompress(logger *slog.Logger, params DecompressParams) (gzip.Header, error) {
	logger.Debug("creating gzip reader")
	in, err := gzip.NewReader(params.Input)
	if err != nil {
		logger.Error("failed to create gzip reader", slog.String("errorMessage", err.Error()))
		return gzip.Header{}, err
	}
	logger.Debug("received gzip header", slog.Any("header", in.Header))
	logger.Debug("reading gzipped data from input decompressing and writing to output")
	if err := util.ProcessStreams(logger, in, params.Output); err != nil {
		logger.Error("failed to read gzip data into output", slog.String("errorMessage", err.Error()))
		return in.Header, err
	}
	logger.Debug("closing gzip reader")
	if err := in.Close(); err != nil {
		logger.Error("failed to close gzip reader", slog.String("errorMessage", err.Error()))
		return in.Header, err
	}
	return in.Header, nil
}
