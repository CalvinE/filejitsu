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

func NewGZIPWriter(logger *slog.Logger, output io.Writer, compressionLevel int, header gzip.Header) (*gzip.Writer, error) {
	logger.Debug("creating gzip writer")
	out, err := gzip.NewWriterLevel(output, compressionLevel)
	if err != nil {
		logger.Debug("failed to create new gzip writer with compression level",
			slog.Int("compressionLevel", compressionLevel),
			slog.String("errorMessage", err.Error()),
		)
		return nil, err
	}
	if len(header.Comment) > 0 {
		logger.Debug("adding header comment", slog.String("comment", header.Comment))
		out.Header.Comment = header.Comment
	}
	if len(header.Name) > 0 {
		logger.Debug("adding header name", slog.String("name", header.Name))
		out.Header.Name = header.Name
	}
	if len(header.Extra) > 0 {
		logger.Debug("adding header extra data", slog.Any("extraData", header.Extra))
		out.Header.Extra = header.Extra
	}
	if header.OS > 0 {
		logger.Debug("adding header os", slog.Int("os", int(header.OS)))
		out.Header.OS = header.OS
	}
	if header.ModTime != emptyTime {
		slog.Debug("adding header modTime", slog.Time("modTime", header.ModTime))
		out.Header.ModTime = header.ModTime
	}
	return out, nil
}

func Compress(logger *slog.Logger, input io.Reader, output *gzip.Writer) error {
	logger.Debug("writing input to gzip writer")
	if err := util.ProcessStreams(logger, input, output); err != nil {
		logger.Error("failed to write input to gzip output stream", slog.String("errorMessage", err.Error()))
		return err
	}
	logger.Debug("closing gzip writer")
	if err := output.Flush(); err != nil {
		logger.Error("failed to flush gzip writer", slog.String("errorMessage", err.Error()))
		return err
	}
	if err := output.Close(); err != nil {
		logger.Error("failed to close gzip writer", slog.String("errorMessage", err.Error()))
		return err
	}
	return nil
}

func NewGZIPReader(logger *slog.Logger, input io.Reader) (*gzip.Reader, gzip.Header, error) {
	logger.Debug("creating gzip reader")
	in, err := gzip.NewReader(input)
	if err != nil {
		logger.Error("failed to create gzip reader", slog.String("errorMessage", err.Error()))
		return nil, gzip.Header{}, err
	}
	logger.Debug("received gzip header", slog.Any("header", in.Header))
	return in, in.Header, nil
}

func Decompress(logger *slog.Logger, input *gzip.Reader, output io.Writer) error {
	logger.Debug("reading gzipped data from input decompressing and writing to output")
	if err := util.ProcessStreams(logger, input, output); err != nil {
		logger.Error("failed to read gzip data into output", slog.String("errorMessage", err.Error()))
		return err
	}
	logger.Debug("closing gzip reader")
	if err := input.Close(); err != nil {
		logger.Error("failed to close gzip reader", slog.String("errorMessage", err.Error()))
		return err
	}
	return nil
}
