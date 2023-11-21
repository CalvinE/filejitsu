package util

import (
	"io"

	"log/slog"
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

func ReaderReadAll(logger *slog.Logger, r io.Reader) ([]byte, error) {
	buffer := make([]byte, 128)
	output := make([]byte, 0)
	bytesRead := 0
	done := false
	for !done {
		rn, err := r.Read(buffer)
		bytesRead += rn
		output = append(output, buffer[:rn]...)
		if err != nil {
			if err == io.EOF {
				// we are done here
				done = true
				continue
			} else {
				logger.Error("failed to read data from input buffer", slog.String("errorMessage", err.Error()))
				return nil, err
			}
		}
		if rn == 0 {
			logger.Debug("no data was read... assume we have read it all?")
			done = true
			continue
		}
	}
	logger.Debug("read data from input", slog.Int("bytesRead", bytesRead))
	return output, nil
}

func AttemptToClose(logger *slog.Logger, input interface{}) error {
	logger.Debug("attempting to close input")
	readCloser, ok := input.(io.Closer)
	if ok {
		err := readCloser.Close()
		if err != nil {
			logger.Error("failed to close input steam", slog.String("errorMessage", err.Error()))
			return err
		}
		logger.Debug("input was a closer")
		return nil
	}
	logger.Debug("input was not a closer")
	return nil
}
