package tar

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/calvine/filejitsu/gzip"
)

func cleanTarTest(t *testing.T, tarFilePath, tarUnpackagePath string) error {
	hadError := false
	var deleteTarFileErr error
	if err := os.Remove(tarFilePath); err != nil {
		hadError = true
		t.Errorf("failed to clean up created tar file: %s - %s", tarFilePath, err.Error())
		deleteTarFileErr = err
	}
	var deleteTarUnpackageErr error
	if err := os.RemoveAll(tarUnpackagePath); err != nil {
		hadError = true
		t.Errorf("failed to clean up created tar file: %s - %s", tarFilePath, err.Error())
		deleteTarUnpackageErr = err
	}
	if hadError {
		err := errors.New("failed to delete at least one test artifact")
		if deleteTarFileErr != nil {
			err = fmt.Errorf("%w: %w", err, deleteTarFileErr)
		}
		if deleteTarUnpackageErr != nil {
			err = fmt.Errorf("%w: %w", err, deleteTarUnpackageErr)
		}
		return err
	}
	return nil
}

func TestTarPackageRoundTrip(t *testing.T) {
	tarPath := "/home/calvin/fjt.tar.gz.enc"
	outputPath := "/home/calvin/fjtdest"
	defer cleanTarTest(t, tarPath, outputPath)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug.Level(),
	}))
	passphrase := []byte("This is my good password")
	output, err := os.OpenFile(tarPath, os.O_CREATE|os.O_RDWR, 0644) // os.Open(tarPath)
	if err != nil {
		t.Errorf("failed to open destination file: %v", err)
		return
	}
	err = TarPackage(logger, TarPackageParams{
		InputPaths: []string{"/home/calvin/code/filejitsu/test_files"},
		UseGzip:    true,
		GZIPOptions: GZIPOptions{
			CompressionLevel: gzip.DefaultCompression,
		},
		UseEncryption: true,
		EncryptionOptions: EncryptionOptions{
			Passphrase: passphrase,
		},
		Output: output,
	})
	if err != nil {
		t.Errorf("failed to write tar output: %v", err)
		return
	}
	// if err := output.Close(); err != nil {
	// 	t.Errorf("failed to close output tar: %v", err)
	// 	return
	// }
	inputTar, err := os.Open(tarPath)
	if err != nil {
		t.Errorf("failed to open tar for unpack: %v", err)
		return
	}
	err = TarUnpackage(logger, TarUnpackageParams{
		Input:         inputTar,
		OutputPath:    outputPath,
		UseGzip:       true,
		UseEncryption: true,
		EncryptionOptions: EncryptionOptions{
			Passphrase: passphrase,
		},
	})
	if err != nil {
		t.Errorf("failed to unpackage tar: %v", err)
		return
	}
}
