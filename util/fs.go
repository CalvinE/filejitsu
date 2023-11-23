package util

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
)

func MakeAllDirIfNotExists(logger *slog.Logger, path string, perm fs.FileMode) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// make it
			if err := os.MkdirAll(path, perm); err != nil {
				logger.Error("failed to make dir all", slog.String("path", path), slog.String("errorMessage", err.Error()))
				return err
			}
		}
		logger.Error("failed to perform stat on path", slog.String("path", path), slog.String("errorMessage", err.Error()))
		return err
	}
	if !info.IsDir() {
		errMsg := "path exists and is not a directory"
		logger.Error(errMsg, slog.String("path", path))
		return errors.New(errMsg)
	}
	return nil
}
