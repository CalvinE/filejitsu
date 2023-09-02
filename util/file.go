package util

import (
	"errors"
	"fmt"
	"os"
)

func OpenFile(path string, flag int, mode os.FileMode) (*os.File, error) {
	// open file reader
	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get stat on path: %w", err)
	}
	if stat.IsDir() {
		return nil, errors.New("path is a directory not a file")
	}
	f, err := os.OpenFile(path, flag, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return f, nil
}
