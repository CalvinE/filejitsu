package util

import (
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"io"

	"log/slog"
)

func Sha512HashData(logger *slog.Logger, data io.Reader) (string, error) {
	hasher := sha512.New()
	return hashData(logger, data, hasher)
}

func hashData(logger *slog.Logger, data io.Reader, hasher hash.Hash) (string, error) {
	sha512.New512_256()
	logger.Debug("creating buffer for file data")
	buffer := make([]byte, 512)
	bytesRead := 0
	// bytesWritten := 0
	for {
		br, err := data.Read(buffer)
		if err != nil {
			if err == io.EOF {
				// we are done here
				logger.Debug("hit EOF", slog.Int("bytesRead", bytesRead))
				break
			} else {
				logger.Error("failed to read data from input reader", slog.String("errorMessage", err.Error()))
				return "", err
			}
		}
		if br > 0 {
			bytesRead += br
			/*bw*/ _, err := hasher.Write(buffer[:br])
			if err != nil {
				logger.Error("failed to write reader data to hasher", slog.String("errorMessage", err.Error()))
				return "", err
			}
			// bytesWritten += bw
		}
	}
	hashString := hex.EncodeToString(hasher.Sum(nil))
	return hashString, nil
}
