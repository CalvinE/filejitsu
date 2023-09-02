package util

import (
	"io"

	"golang.org/x/exp/slog"
)

// ProcessStreams writes data from an io.Reader to and io.Writer
func ProcessStreams(logger *slog.Logger, input io.Reader, output io.Writer) error {
	inputBuffer := make([]byte, 64)
	bytesRead, bytesWritten := 0, 0
	done := false
	for !done {
		rn, err := input.Read(inputBuffer)
		bytesRead += rn
		if err != nil {
			if err == io.EOF {
				// we are done here
				done = true
			} else {
				logger.Error("failed to read data from input buffer", slog.String("errorMessage", err.Error()))
				return err
			}
		}
		if rn > 0 {
			wn, err := output.Write(inputBuffer[:rn])
			bytesWritten += wn
			if err != nil {
				logger.Error("failed to write to output stream", slog.String("errorMessage", err.Error()))
				return err
			}
		}
	}
	logger.Debug("done processing stream", slog.Int("bytesWritten", bytesWritten), slog.Int("bytesRead", bytesRead))
	return nil
}
