package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"io"

	"log/slog"

	"github.com/calvine/filejitsu/util"
)

func NewAESDecryptionReader(logger *slog.Logger, input io.Reader, passphrase []byte) (*cipher.StreamReader, error) {
	iv := make([]byte, aes.BlockSize)
	n, err := input.Read(iv)
	if err != nil {
		logger.Error("failed to read iv from input", slog.String("errorMessage", err.Error()))
		return nil, err
	}
	ivLen := len(iv)
	if n != ivLen {
		err := fmt.Errorf("iv read from input not expected length: got %d - expected %d", n, ivLen)
		logger.Error("failed to read iv from input", slog.String("errorMessage", err.Error()), slog.Int("got", n), slog.Int("expected", ivLen))
		return nil, err
	}
	hashedPassphrase := sha256.Sum256(passphrase)
	block, err := aes.NewCipher(hashedPassphrase[:])
	if err != nil {
		logger.Error("failed to create new cipher block", slog.String("errorMessage", err.Error()))
		return nil, err
	}
	stream := cipher.NewOFB(block, iv)
	cipherStream := cipher.StreamReader{S: stream, R: input}
	return &cipherStream, nil
}

func Decrypt(logger *slog.Logger, cipherStream *cipher.StreamReader, output io.Writer) error {
	// iv := make([]byte, aes.BlockSize)
	// n, err := params.Input.Read(iv)
	// if err != nil {
	// 	logger.Error("failed to read iv from input", slog.String("errorMessage", err.Error()))
	// 	return err
	// }
	// ivLen := len(iv)
	// if n != ivLen {
	// 	err := fmt.Errorf("iv read from input not expected length: got %d - expected %d", n, ivLen)
	// 	logger.Error("failed to read iv from input", slog.String("errorMessage", err.Error()), slog.Int("got", n), slog.Int("expected", ivLen))
	// 	return err
	// }
	// hashedPassphrase := sha256.Sum256(params.Passphrase)
	// block, err := aes.NewCipher(hashedPassphrase[:])
	// if err != nil {
	// 	logger.Error("failed to create new cipher block", slog.String("errorMessage", err.Error()))
	// 	return err
	// }
	// stream := cipher.NewOFB(block, iv)
	// cipherStream := cipher.StreamReader{S: stream, R: params.Input}
	err := util.ProcessStreams(logger, cipherStream, output)
	if err != nil {
		logger.Error("failed to decrypt data", slog.String("errorMessage", err.Error()))
		return err
	}
	logger.Debug("done decrypting input")
	return nil
}
